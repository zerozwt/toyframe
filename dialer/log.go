package dialer

import (
	"log"
	"sync/atomic"
)

var g_logger atomic.Value

func logger() *log.Logger {
	return g_logger.Load().(*log.Logger)
}

func SetLogWriter(logger *log.Logger) {
	g_logger.Store(logger)
}
