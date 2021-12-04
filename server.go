package toyframe

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var ErrDuplicateMethod error = errors.New("duplicate method")

type Server interface {
	Register(method string, cb Handler) error
	AddListener(lis net.Listener) Server
	Run() error
	Close() error
}

type Handler func(*Context) error

func NewServer() Server {
	return &server{}
}

type server struct {
	sync.RWMutex

	route map[string]Handler
	lis   []net.Listener

	run      int32
	closed   int32
	close_ch chan struct{}
}

func (s *server) Register(method string, cb Handler) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.route[method]; ok {
		return ErrDuplicateMethod
	}
	s.route[method] = cb
	return nil
}

func (s *server) AddListener(lis net.Listener) Server {
	s.Lock()
	defer s.Unlock()
	s.lis = append(s.lis, lis)
	return s
}

func (s *server) Run() error {
	if !atomic.CompareAndSwapInt32(&(s.run), 0, 1) {
		return nil
	}

	s.RLock()
	for idx := range s.lis {
		go s.serve(s.lis[idx])
	}
	s.RUnlock()

	<-s.close_ch
	return nil
}

func (s *server) Close() error {
	if !atomic.CompareAndSwapInt32(&(s.closed), 0, 1) {
		return nil
	}
	s.Lock()
	defer s.Unlock()

	for _, lis := range s.lis {
		lis.Close()
	}
	s.lis = nil
	close(s.close_ch)
	return nil
}

func (s *server) serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		go s.serveContext(conn)
	}
}

func (s *server) serveContext(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			logger().Printf("serve context panic: %v", err)
			return
		}
	}()

	m_len := [1]byte{}
	if _, err := io.ReadFull(conn, m_len[:]); err != nil {
		logger().Printf("read method length from %v failed: %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}
	method_data := make([]byte, m_len[0])
	if _, err := io.ReadFull(conn, method_data); err != nil {
		logger().Printf("read method from %v failed: %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}

	method := string(method_data)
	s.RLock()
	handler, ok := s.route[method]
	s.RUnlock()
	if !ok {
		logger().Printf("method %s handler not found", method)
		conn.Close()
		return
	}

	ctx := newContext(conn)
	ctx.Method = method
	defer ctx.Close()
	handler(ctx)
}
