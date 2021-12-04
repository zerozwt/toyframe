package dialer

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/yamux"
	"github.com/zerozwt/toyframe/conn"
)

func multiplexDial(dial DialFunc) DialFunc {
	obj := &multiplexDialer{
		dial:  dial,
		cache: make(map[mdCacheKey]*multiplexSession),
	}
	return obj.Dial
}

type mdCacheKey [2]string // [network, address]

type multiplexDialer struct {
	sync.RWMutex
	dial  DialFunc
	cache map[mdCacheKey]*multiplexSession
}

func (d *multiplexDialer) Dial(network, addr string) (net.Conn, error) {
	sess := d.getSession(network, addr)
	var err error

	if sess == nil {
		sess, err = d.createSession(network, addr)
		if err != nil {
			return nil, err
		}
	}

	socket, err := sess.dial()
	if err != nil {
		// serve side session may have gone away, try creating a new session
		d.deleteSession(network, addr, sess)
		sess, err = d.createSession(network, addr)
		if err != nil {
			return nil, err
		}
		socket, err = sess.dial()
	}

	return socket, err
}

func (d *multiplexDialer) getSession(network, addr string) *multiplexSession {
	d.RLock()
	defer d.RUnlock()
	key := mdCacheKey{network, addr}
	if sess, ok := d.cache[key]; ok {
		return sess
	}
	return nil
}

func (d *multiplexDialer) createSession(network, addr string) (*multiplexSession, error) {
	d.Lock()
	defer d.Unlock()
	key := mdCacheKey{network, addr}
	if sess, ok := d.cache[key]; ok {
		return sess, nil
	}

	socket, err := d.dial(network, addr)
	if err != nil {
		return nil, err
	}
	sess := newMultiplexSession(d, socket, network, addr)
	d.cache[key] = sess
	return sess, nil
}

func (d *multiplexDialer) deleteSession(network, addr string, session *multiplexSession) {
	d.Lock()
	defer d.Unlock()
	key := mdCacheKey{network, addr}
	if sess, ok := d.cache[key]; ok && sess == session {
		delete(d.cache, key)
		session.closeIfEmpty()
	}
}

type multiplexSession struct {
	sync.Mutex
	sess *yamux.Session
	host *multiplexDialer

	network string
	addr    string

	conn_id  uint32
	conn_map map[uint32]net.Conn
}

func newMultiplexSession(host *multiplexDialer, conn net.Conn, network, addr string) *multiplexSession {
	conf := yamux.DefaultConfig()
	conf.LogOutput = nil
	conf.Logger = logger()

	sess, _ := yamux.Client(conn, conf)

	return &multiplexSession{
		sess:     sess,
		host:     host,
		network:  network,
		addr:     addr,
		conn_map: make(map[uint32]net.Conn),
	}
}

func (s *multiplexSession) dial() (net.Conn, error) {
	socket, err := s.sess.Open()
	if err != nil {
		return nil, err
	}
	socket = conn.NewMultiplexConn(s, s.acquireID(socket), socket)
	return socket, nil
}

func (s *multiplexSession) acquireID(conn net.Conn) uint32 {
	s.Lock()
	defer s.Unlock()
	ret := atomic.AddUint32(&(s.conn_id), 1)
	s.conn_map[ret] = conn
	return ret
}

func (s *multiplexSession) RemoveConn(id uint32) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.conn_map[id]; ok {
		delete(s.conn_map, id)
		if len(s.conn_map) == 0 && s.host.getSession(s.network, s.addr) != s {
			s.sess.Close()
		}
	}
}

func (s *multiplexSession) closeIfEmpty() {
	s.Lock()
	defer s.Unlock()
	if len(s.conn_map) == 0 {
		s.sess.Close()
	}
}
