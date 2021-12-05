package toyframe

import (
	"errors"
	"io"
)

var ErrInterrupted error = errors.New("interrupted")

func DoWithInterruptor(job func(), interrupt_ch chan struct{}) error {
	var panic_value interface{}
	result_ch := make(chan struct{})
	go func() {
		defer func() {
			panic_value = recover()
			close(result_ch)
		}()
		job()
	}()
	select {
	case <-result_ch:
		if panic_value != nil {
			panic(panic_value)
		}
		return nil
	case <-interrupt_ch:
		return ErrInterrupted
	}
}

func InterruptableReader(reader io.Reader, interrupt_ch chan struct{}) io.Reader {
	return &interruptableReader{reader: reader, ich: interrupt_ch}
}

func InterruptableWriter(writer io.WriteCloser, interrupt_ch chan struct{}) io.WriteCloser {
	return &interruptableWriteCloser{writer: writer, ich: interrupt_ch}
}

type interruptableReader struct {
	reader io.Reader
	ich    chan struct{}
}

func (r *interruptableReader) Read(buf []byte) (n int, err error) {
	if err2 := DoWithInterruptor(func() { n, err = r.reader.Read(buf) }, r.ich); err2 != nil {
		return 0, err2
	}
	return
}

type interruptableWriteCloser struct {
	writer io.WriteCloser
	ich    chan struct{}
}

func (w *interruptableWriteCloser) Write(buf []byte) (n int, err error) {
	if err2 := DoWithInterruptor(func() { n, err = w.writer.Write(buf) }, w.ich); err2 != nil {
		return 0, err2
	}
	return
}

func (w *interruptableWriteCloser) Close() error {
	return w.writer.Close()
}

func readSimpleString(reader io.Reader) (string, error) {
	len := [1]byte{}
	if _, err := io.ReadFull(reader, len[:]); err != nil {
		return "", err
	}
	data := make([]byte, len[0])
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func writeSimpleString(writer io.Writer, data string) error {
	buf := make([]byte, 1, 1+len(data))
	buf[0] = byte(len(data))
	for _, ch := range data {
		buf = append(buf, byte(ch))
	}

	_, err := writer.Write(buf)
	return err
}
