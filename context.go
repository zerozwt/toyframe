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

func Call(network, addr, method string, dial dialer.DialFunc, params ...msgp.MarshalSizer) (*Context, error) {
	if len(method) > 0xFF {
		// method == "" is allowed
		return nil, ErrMethodTooLong
	}

	// dial remote addr and create context
	conn, err := dial(network, addr)
	if err != nil {
		return nil, err
	}
	ctx := newContext(conn)

	// send method message
	if err = writeSimpleString(ctx.writer, method); err != nil {
		ctx.Close()
		return nil, err
	}

	// send params asynchronously
	send_param_ch := make(chan error, 1)
	go func() {
		for _, item := range params {
			if err := ctx.WriteObj(item); err != nil {
				send_param_ch <- err
				return
			}
		}
		send_param_ch <- nil
	}()

	// read call method result
	result, err := readSimpleString(ctx.reader)
	if err != nil {
		ctx.Close()
		return nil, err
	}
	if len(result) > 0 {
		ctx.Close()
		return nil, errors.New(result)
	}

	// read send param result
	err = <-send_param_ch
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

func readSimpleString(reader io.Reader) (string, error) {
	len := [1]byte{}
	if _, err := io.ReadFull(reader, len[:]); err != nil {
		return "", err
	}
	data := make([]byte, len[0])
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func writeSimpleString(writer io.Writer, data string) error {
	buf := make([]byte, 1, 1+len(data))
	buf[0] = byte(len(data))
	for _, ch := range data {
		buf = append(buf, byte(ch))
	}

	_, err := writer.Write(buf)
	return err
}
