package dialer

import (
	"crypto/tls"
	"net"
)

func tlsDial(dial DialFunc, conf *tls.Config) DialFunc {
	return func(network, addr string) (net.Conn, error) {
		conn, err := dial(network, addr)
		if err != nil {
			return conn, err
		}
		return tls.Client(conn, conf), nil
	}
}
