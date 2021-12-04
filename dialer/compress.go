package dialer

import (
	"net"

	"github.com/zerozwt/toyframe/conn"
)

func brotliDial(dial DialFunc) DialFunc {
	return func(network, addr string) (net.Conn, error) {
		socket, err := dial(network, addr)
		if err != nil {
			return socket, err
		}
		return conn.NewBrotliConn(socket), err
	}
}
