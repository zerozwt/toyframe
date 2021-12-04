package listener

import (
	"crypto/tls"
	"net"
)

type Builder struct {
	lis net.Listener
}

func B(lis net.Listener) *Builder {
	return &Builder{lis: lis}
}

func (b *Builder) Build() net.Listener {
	return &writeLockListener{lis: b.lis}
}

func (b *Builder) WithTls(conf *tls.Config) *Builder {
	return &Builder{lis: &tlsListener{lis: b.lis, conf: conf}}
}

func (b *Builder) WithBrotli() *Builder {
	return &Builder{lis: &brotliListener{lis: b.lis}}
}

func (b *Builder) WithBitReverse() *Builder {
	return &Builder{lis: &bitReverseListener{lis: b.lis}}
}

func (b *Builder) WithMultiplex() *Builder {
	return &Builder{lis: newMultiplexListener(b.lis)}
}
