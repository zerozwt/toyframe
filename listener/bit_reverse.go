package listener

import (
	"net"

	"github.com/zerozwt/toyframe/conn"
)

type bitReverseListener struct {
	lis net.Listener
}

func (l *bitReverseListener) Accept() (net.Conn, error) {
	socket, err := l.lis.Accept()
	if err != nil {
		return nil, err
	}
	return conn.NewBitReverseConn(socket), nil
}

func (l *bitReverseListener) Close() error {
	return l.lis.Close()
}

func (l *bitReverseListener) Addr() net.Addr {
	return l.lis.Addr()
}
