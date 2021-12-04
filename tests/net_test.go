package tests

import (
	"bytes"
	"crypto/tls"
	"net"
	"os"
	"testing"

	"github.com/zerozwt/toyframe"
	"github.com/zerozwt/toyframe/dialer"
	"github.com/zerozwt/toyframe/listener"
)

func testServerLogic(t *testing.T, lis net.Listener, network, addr string, dial dialer.DialFunc) {
	server := toyframe.NewServer()

	server.AddListener(lis)
	server.Register("ping", func(ctx *toyframe.Context) error {
		msg := TestMsg{}
		if err := ctx.ReadObj(&msg); err != nil {
			t.Errorf("server read obj failed: %v\n", err)
			return nil
		}
		if err := ctx.WriteObj(&msg); err != nil {
			t.Errorf("server write obj failed: %v\n", err)
			return nil
		}
		return nil
	})

	server_close_ch := make(chan struct{})
	go func() {
		server.Run()
		close(server_close_ch)
	}()

	msg := TestMsg{
		ID:   123,
		Name: "hello",
		Data: []byte("world"),
	}
	ctx, err := toyframe.Call(network, addr, "ping", dial, &msg)
	if err != nil {
		t.Errorf("call to server failed: %v\n", err)
		return
	}
	msg2 := TestMsg{}
	if err := ctx.ReadObj(&msg2); err != nil {
		t.Errorf("read msg from server failed: %v\n", err)
		return
	}

	if msg.ID != msg2.ID || msg.Name != msg2.Name || !bytes.Equal(msg.Data, msg2.Data) {
		t.Errorf("ping pong msg not exactly same: msg=%v msg2=%v", msg, msg2)
		return
	}

	ctx.Close()
	server.Close()
	<-server_close_ch
}

func TestNet_BitReverse(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	lis, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithBitReverse().Build()
	dial := dialer.B(net.Dial).WithBitReverse().Build()

	testServerLogic(t, lis, "tcp", "localhost:7777", dial)
}

func TestNet_Brotli(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	lis, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithBrotli().Build()
	dial := dialer.B(net.Dial).WithBrotli().Build()

	testServerLogic(t, lis, "tcp", "localhost:7777", dial)
}

func TestNet_Tls(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	cert, _ := tls.X509KeyPair(tlsPem, tlsKey)

	server_conf := &tls.Config{Certificates: []tls.Certificate{cert}}
	client_conf := &tls.Config{InsecureSkipVerify: true}

	lis, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithTls(server_conf).Build()
	dial := dialer.B(net.Dial).WithTls(client_conf).Build()

	testServerLogic(t, lis, "tcp", "localhost:7777", dial)
}

func TestNet_MultiplexBasic(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	lis, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithMultiplex().Build()
	dial := dialer.B(net.Dial).WithMultiplex().Build()

	testServerLogic(t, lis, "tcp", "localhost:7777", dial)
}

func TestNet_All(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	cert, _ := tls.X509KeyPair(tlsPem, tlsKey)

	server_conf := &tls.Config{Certificates: []tls.Certificate{cert}}
	client_conf := &tls.Config{InsecureSkipVerify: true}

	lis, err := net.Listen("tcp", "localhost:7777")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithBitReverse().WithTls(server_conf).WithMultiplex().WithBrotli().Build()
	dial := dialer.B(net.Dial).WithBitReverse().WithTls(client_conf).WithMultiplex().WithBrotli().Build()

	testServerLogic(t, lis, "tcp", "localhost:7777", dial)
}
