package conn

import (
	"net"
	"time"
)

type BitReverseConn struct {
	conn net.Conn
}

func NewBitReverseConn(conn net.Conn) net.Conn {
	return &BitReverseConn{conn: conn}
}

func (c *BitReverseConn) Read(b []byte) (int, error) {
	n, err := c.conn.Read(b)
	for i := 0; i < n; i++ {
		b[i] = b[i] ^ 0xFF
	}
	return n, err
}

func (c *BitReverseConn) Write(b []byte) (int, error) {
	for i := range b {
		b[i] = b[i] ^ 0xFF
	}
	n, err := c.conn.Write(b)
	for i := n; i < len(b); i++ {
		b[i] = b[i] ^ 0xFF
	}
	return n, err
}

func (c *BitReverseConn) Close() error {
	return c.conn.Close()
}

func (c *BitReverseConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *BitReverseConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *BitReverseConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *BitReverseConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *BitReverseConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
