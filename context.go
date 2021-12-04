package toyframe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync/atomic"

	"github.com/tinylib/msgp/msgp"
	"github.com/zerozwt/toyframe/dialer"
)

var ErrMethodTooLong error = errors.New("method too long")

type Context struct {
	reader io.Reader
	writer io.WriteCloser

	Method string

	conn   net.Conn
	closed int32
}

func Call(network, addr, method string, dial dialer.DialFunc) (*Context, error) {
	if len(method) > 0xFF {
		// method == "" is allowed
		return nil, ErrMethodTooLong
	}

	conn, err := dial(network, addr)
	if err != nil {
		return nil, err
	}
	ctx := newContext(conn)

	buf := make([]byte, 1, 1+len(method))
	buf[0] = byte(len(method))
	for _, ch := range method {
		buf = append(buf, byte(ch))
	}

	_, err = ctx.writer.Write(buf)
	if err != nil {
		ctx.Close()
		return nil, err
	}

	return ctx, nil
}

func (c *Context) Close() error {
	if atomic.CompareAndSwapInt32(&(c.closed), 0, 1) {
		return c.conn.Close()
	}
	return nil
}

func (c *Context) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Context) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Context) Reader() io.Reader {
	return c.reader
}

func (c *Context) Writer() io.Writer {
	return c.writer
}

func (c *Context) ReadObj(obj msgp.Unmarshaler) error {
	var len uint32
	if err := binary.Read(c.reader, binary.BigEndian, &len); err != nil {
		return err
	}
	buf := make([]byte, len)
	if _, err := io.ReadFull(c.reader, buf); err != nil {
		return err
	}
	_, err := obj.UnmarshalMsg(buf)
	return err
}

func (c *Context) WriteObj(obj msgp.MarshalSizer) error {
	buf := make([]byte, 4, 4+obj.Msgsize())
	buf, err := obj.MarshalMsg(buf)
	if err != nil {
		return err
	}
	if err = binary.Write(bytes.NewBuffer(buf[:4]), binary.BigEndian, uint32(len(buf)-4)); err != nil {
		return err
	}
	_, err = c.writer.Write(buf)
	return err
}

func newContext(conn net.Conn) *Context {
	ctx := &Context{
		reader: conn,
		conn:   conn,
	}
	ctx.writer = &contextWriteCloser{
		conn: conn,
		ctx:  ctx,
	}
	return ctx
}

type contextWriteCloser struct {
	conn io.Writer
	ctx  *Context
}

func (wc *contextWriteCloser) Write(buf []byte) (int, error) {
	return wc.conn.Write(buf)
}

func (wc *contextWriteCloser) Close() error {
	return wc.ctx.Close()
}
