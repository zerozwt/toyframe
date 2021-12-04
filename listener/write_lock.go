package listener

import (
	"net"

	"github.com/zerozwt/toyframe/conn"
)

type writeLockListener struct {
	lis net.Listener
}

func (l *writeLockListener) Accept() (net.Conn, error) {
	socket, err := l.lis.Accept()
	if err != nil {
		return nil, err
	}
	return conn.NewWriteLockConn(socket), nil
}

func (l *writeLockListener) Close() error {
	return l.lis.Close()
}

func (l *writeLockListener) Addr() net.Addr {
	return l.lis.Addr()
}
