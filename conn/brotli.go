package conn

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/andybalholm/brotli"
)

type BrotliConn struct {
	conn net.Conn
	r    *brotli.Reader
	w    *brotli.Writer

	closed int32
}

func NewBrotliConn(conn net.Conn) net.Conn {
	return &BrotliConn{
		conn: conn,
		r:    brotli.NewReader(conn),
		w:    brotli.NewWriter(conn),
	}
}

func (c *BrotliConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *BrotliConn) Write(b []byte) (int, error) {
	n, err := c.w.Write(b)
	if err != nil {
		return n, err
	}
	err = c.w.Flush()
	return n, err
}

func (c *BrotliConn) Close() error {
	if !atomic.CompareAndSwapInt32(&(c.closed), 0, 1) {
		return nil
	}
	c.w.Close()
	return c.conn.Close()
}

func (c *BrotliConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *BrotliConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *BrotliConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *BrotliConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *BrotliConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
