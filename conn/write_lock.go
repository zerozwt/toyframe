package conn

import (
	"net"
	"sync"
	"time"
)

type WriteLockConn struct {
	sync.Mutex
	conn net.Conn
}

func NewWriteLockConn(conn net.Conn) net.Conn {
	return &WriteLockConn{conn: conn}
}

func (c *WriteLockConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *WriteLockConn) Write(b []byte) (int, error) {
	c.Lock()
	defer c.Unlock()
	return c.conn.Write(b)
}

func (c *WriteLockConn) Close() error {
	c.Lock()
	defer c.Unlock()
	return c.conn.Close()
}

func (c *WriteLockConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WriteLockConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *WriteLockConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *WriteLockConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WriteLockConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
