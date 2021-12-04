package conn

import (
	"net"
	"sync/atomic"
	"time"
)

type MultiplexSession interface {
	RemoveConn(id uint32)
}

type MultiplexConn struct {
	id      uint32
	session MultiplexSession
	conn    net.Conn
	closed  int32
}

func NewMultiplexConn(session MultiplexSession, id uint32, conn net.Conn) net.Conn {
	return &MultiplexConn{session: session, id: id, conn: conn}
}

func (c *MultiplexConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *MultiplexConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *MultiplexConn) Close() error {
	if !atomic.CompareAndSwapInt32(&(c.closed), 0, 1) {
		return nil
	}
	defer c.session.RemoveConn(c.id)
	return c.conn.Close()
}

func (c *MultiplexConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *MultiplexConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *MultiplexConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *MultiplexConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *MultiplexConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
