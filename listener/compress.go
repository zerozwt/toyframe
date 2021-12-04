package listener

import (
	"net"

	"github.com/zerozwt/toyframe/conn"
)

type brotliListener struct {
	lis net.Listener
}

func (l *brotliListener) Accept() (net.Conn, error) {
	socket, err := l.lis.Accept()
	if err != nil {
		return nil, err
	}
	return conn.NewBrotliConn(socket), nil
}

func (l *brotliListener) Close() error {
	return l.lis.Close()
}

func (l *brotliListener) Addr() net.Addr {
	return l.lis.Addr()
}
