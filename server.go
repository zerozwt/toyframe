package toyframe

import (
	"errors"
	"fmt"
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
	return &server{
		route:    make(map[string]Handler),
		lis:      make([]net.Listener, 0),
		close_ch: make(chan struct{}),
	}
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

	method, err := readSimpleString(conn)
	if err != nil {
		logger().Printf("read method from %v failed: %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}

	s.RLock()
	handler, ok := s.route[method]
	s.RUnlock()
	if !ok {
		msg := fmt.Sprintf("method %s handler not found", method)
		logger().Println(msg)
		writeSimpleString(conn, msg)
		conn.Close()
		return
	}
	writeSimpleString(conn, "") // empty string means handler successfully found

	ctx := newContext(conn)
	ctx.Method = method
	defer ctx.Close()
	handler(ctx)
}
