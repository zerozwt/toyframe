package toyframe

import (
	"io"
	"log"
	"sync/atomic"

	"github.com/zerozwt/toyframe/dialer"
	"github.com/zerozwt/toyframe/listener"
)

var g_logger atomic.Value

func init() {
	SetLogWriter(NullWriter{})
}

func logger() *log.Logger {
	return g_logger.Load().(*log.Logger)
}

func SetLogWriter(out io.Writer) {
	logger := log.New(out, "toyframe", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	g_logger.Store(logger)
	listener.SetLogWriter(logger)
	dialer.SetLogWriter(logger)
}

type NullWriter struct{}

func (w NullWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
