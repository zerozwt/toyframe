package tests

import (
	"testing"
	"time"

	"github.com/zerozwt/toyframe"
)

type sleeperIO time.Duration

func (s sleeperIO) Read(buf []byte) (int, error) {
	time.Sleep(time.Duration(s))
	return len(buf), nil
}

func (s sleeperIO) Write(buf []byte) (int, error) {
	time.Sleep(time.Duration(s))
	return len(buf), nil
}

func (s sleeperIO) Close() error { return nil }

func timeInterruptor(value time.Duration) chan struct{} {
	ret := make(chan struct{})
	go func() {
		<-time.After(value)
		close(ret)
	}()
	return ret
}

func TestInterrupt(t *testing.T) {
	io := sleeperIO(time.Second * 5)
	reader := toyframe.InterruptableReader(io, timeInterruptor(time.Millisecond*10))
	buf := [8]byte{}

	_, err := reader.Read(buf[:])
	if err != toyframe.ErrInterrupted {
		t.Errorf("Reader not interrupted correctly: %v", err)
		return
	}

	writer := toyframe.InterruptableWriter(io, timeInterruptor(time.Millisecond*10))
	_, err = writer.Write(buf[:])
	if err != toyframe.ErrInterrupted {
		t.Errorf("Writer not interrupted correctly: %v", err)
		return
	}

	io = sleeperIO(time.Microsecond * 5)
	reader = toyframe.InterruptableReader(io, timeInterruptor(time.Second))
	_, err = reader.Read(buf[:])
	if err != nil {
		t.Errorf("Reader not done correctly: %v", err)
		return
	}
	writer = toyframe.InterruptableWriter(io, timeInterruptor(time.Second))
	_, err = writer.Write(buf[:])
	if err != nil {
		t.Errorf("Writer not done correctly: %v", err)
		return
	}
}
