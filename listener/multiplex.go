package listener

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/yamux"
	"github.com/zerozwt/toyframe/conn"
)

var err_closed error = errors.New("closed")

type multiplexListener struct {
	sync.Mutex

	lis   net.Listener
	start int32

	accept_ch chan acceptResult

	close_ch chan struct{}
	closed   int32
}

func newMultiplexListener(lis net.Listener) net.Listener {
	return &multiplexListener{
		lis:       lis,
		accept_ch: make(chan acceptResult),
		close_ch:  make(chan struct{}),
	}
}

type acceptResult struct {
	conn net.Conn
	err  error
}

func (l *multiplexListener) Accept() (net.Conn, error) {
	if atomic.CompareAndSwapInt32(&(l.start), 0, 1) {
		go l.listen()
	}
	select {
	case result := <-l.accept_ch:
		return result.conn, result.err
	case <-l.close_ch:
		return nil, err_closed
	}
}

func (l *multiplexListener) listen() {
	for {
		conn, err := l.lis.Accept()
		if err != nil {
			l.Close()
			return
		}
		go newSession(l).run(conn)
	}
}

func (l *multiplexListener) Close() error {
	if atomic.CompareAndSwapInt32(&(l.closed), 0, 1) {
		err := l.lis.Close()
		close(l.close_ch)
		return err
	}
	return nil
}

func (l *multiplexListener) Addr() net.Addr {
	return l.lis.Addr()
}

type multiplexSession struct {
	sync.Mutex
	sess *yamux.Session
	host *multiplexListener
	gone int32

	conn_id  uint32
	conn_map map[uint32]net.Conn
}

func newSession(host *multiplexListener) *multiplexSession {
	return &multiplexSession{
		host:     host,
		conn_map: make(map[uint32]net.Conn),
	}
}

func (s *multiplexSession) acquireID(conn net.Conn) uint32 {
	s.Lock()
	defer s.Unlock()
	ret := atomic.AddUint32(&s.conn_id, 1)
	s.conn_map[ret] = conn
	return ret
}

func (s *multiplexSession) RemoveConn(id uint32) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.conn_map[id]
	if !ok {
		return
	}
	delete(s.conn_map, id)

	if s.hostClosed() {
		s.goAwayAndClose()
	}
}

func (s *multiplexSession) run(socket net.Conn) {
	conf := yamux.DefaultConfig()
	conf.LogOutput = nil
	conf.Logger = logger()

	s.sess, _ = yamux.Server(socket, conf)

	go func() {
		<-s.host.close_ch
		s.goAway()
	}()

	for {
		if s.hostClosed() {
			return
		}
		socket, err := s.sess.Accept()
		if err != nil {
			s.goAway()
			return
		}
		s.host.accept_ch <- acceptResult{conn: conn.NewMultiplexConn(s, s.acquireID(socket), socket), err: err}
	}
}

func (s *multiplexSession) hostClosed() bool {
	select {
	case <-s.host.close_ch:
		return true
	default:
		return false
	}
}

func (s *multiplexSession) goAway() {
	s.Lock()
	defer s.Unlock()
	s.goAwayAndClose()
}

func (s *multiplexSession) goAwayAndClose() {
	if atomic.CompareAndSwapInt32((&s.gone), 0, 1) {
		s.sess.GoAway()
	}
	if len(s.conn_map) == 0 {
		s.sess.Close()
	}
}
