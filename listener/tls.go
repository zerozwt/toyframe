package listener

import (
	"crypto/tls"
	"net"
)

type tlsListener struct {
	lis  net.Listener
	conf *tls.Config
}

func (l *tlsListener) Accept() (net.Conn, error) {
	conn, err := l.lis.Accept()
	if err != nil {
		return nil, err
	}
	conn = tls.Server(conn, l.conf)
	return conn, nil
}

func (l *tlsListener) Close() error {
	return l.lis.Close()
}

func (l *tlsListener) Addr() net.Addr {
	return l.lis.Addr()
}
