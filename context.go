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

	onClose []func()

	conn   net.Conn
	closed int32
}

func NewContext(conn net.Conn) *Context {
	return newContext(conn)
}

func Call(network, addr, method string, dial dialer.DialFunc, params ...msgp.MarshalSizer) (*Context, error) {
	return CallWithInterruptor(network, addr, method, dial, nil, params...)
}

func CallWithInterruptor(network, addr, method string, dial dialer.DialFunc, ich chan struct{}, params ...msgp.MarshalSizer) (*Context, error) {
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
	ctx.SetInterruptor(ich)
	defer ctx.SetInterruptor(nil)

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
		defer func() {
			for _, cb := range c.onClose {
				func() {
					defer func() {
						if err := recover(); err != nil {
							logger().Printf("panic in context's close callbacks: %v", err)
						}
					}()
					cb()
				}()
			}
		}()
		return c.conn.Close()
	}
	return nil
}

func (c *Context) AddCloseHandler(cb func()) {
	c.onClose = append(c.onClose, cb)
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
	if err = binary.Write(bytes.NewBuffer(buf[:0]), binary.BigEndian, uint32(len(buf)-4)); err != nil {
		return err
	}
	_, err = c.writer.Write(buf)
	return err
}

func (c *Context) SetInterruptor(ich chan struct{}) {
	if ich == nil {
		c.reader = c.conn
		c.writer = &contextWriteCloser{
			conn: c.conn,
			ctx:  c,
		}
		return
	}
	c.reader = InterruptableReader(c.conn, ich)
	c.writer = InterruptableWriter(&contextWriteCloser{
		conn: c.conn,
		ctx:  c,
	}, ich)
}

func newContext(conn net.Conn) *Context {
	ctx := &Context{
		reader: conn,
		conn:   conn,

		onClose: make([]func(), 0),
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
