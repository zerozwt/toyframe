package dialer

import (
	"crypto/tls"
	"net"
)

type DialFunc func(network, addr string) (net.Conn, error)

type Builder struct {
	dial DialFunc
}

func B(dial DialFunc) *Builder {
	return &Builder{dial: dial}
}

func (b *Builder) Build() DialFunc {
	return writeLockDial(b.dial)
}

func (b *Builder) WithBitReverse() *Builder {
	return &Builder{dial: bitReverseDial(b.dial)}
}

func (b *Builder) WithBrotli() *Builder {
	return &Builder{dial: brotliDial(b.dial)}
}

func (b *Builder) WithTls(conf *tls.Config) *Builder {
	return &Builder{dial: tlsDial(b.dial, conf)}
}

func (b *Builder) WithMultiplex() *Builder {
	return &Builder{dial: multiplexDial(b.dial)}
}
