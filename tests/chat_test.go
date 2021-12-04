package tests

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/zerozwt/toyframe"
	"github.com/zerozwt/toyframe/dialer"
	"github.com/zerozwt/toyframe/listener"
)

type mailCenter struct {
	sync.Mutex
	box   map[string]chan ChatMsg_Send
	cache map[string][]ChatMsg_Send
}

var gMailCenter *mailCenter = &mailCenter{
	box:   make(map[string]chan ChatMsg_Send),
	cache: make(map[string][]ChatMsg_Send),
}

func (c *mailCenter) register(name string, mailbox chan ChatMsg_Send) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.box[name]; ok {
		return
	}
	if list, ok := c.cache[name]; ok {
		for _, item := range list {
			mailbox <- item
		}
		delete(c.cache, name)
	}
	c.box[name] = mailbox
}

func (c *mailCenter) unregister(name string) {
	c.Lock()
	defer c.Unlock()
	if box, ok := c.box[name]; ok {
		delete(c.box, name)
		close(box)
		return
	}
}

func (c *mailCenter) send(msg ChatMsg_Send) {
	c.Lock()
	defer c.Unlock()
	if box, ok := c.box[msg.Reciever]; ok {
		box <- msg
	} else {
		c.cache[msg.Reciever] = append(c.cache[msg.Reciever], msg)
	}
}

func loginHandler(ctx *toyframe.Context, t *testing.T) error {
	// recv login msg
	login := ChatMsg_Login{}
	if err := ctx.ReadObj(&login); err != nil {
		t.Errorf("read login msg failed: %v", err)
		return err
	}

	// register mailbox
	mailbox := make(chan ChatMsg_Send, 16)
	gMailCenter.register(login.Name, mailbox)

	// recv msg
	for msg := range mailbox {
		if err := ctx.WriteObj(&msg); err != nil {
			t.Errorf("send chat msg failed: %v", err)
			return err
		}
	}
	return nil
}

func sendHandler(ctx *toyframe.Context, t *testing.T) error {
	for {
		msg := ChatMsg_Send{}
		if err := ctx.ReadObj(&msg); err != nil {
			if err == io.EOF {
				return nil
			}
			t.Errorf("read msg failed: %v", err)
			return err
		}
		gMailCenter.send(msg)
	}
}

func logoutHandler(ctx *toyframe.Context, t *testing.T) error {
	// recv login msg
	logout := ChatMsg_Login{}
	if err := ctx.ReadObj(&logout); err != nil {
		t.Errorf("read logout msg failed: %v", err)
		return err
	}
	gMailCenter.unregister(logout.Name)
	return nil
}

func setupChatServer(t *testing.T, listeners ...net.Listener) (toyframe.Server, error) {
	server := toyframe.NewServer()

	for _, lis := range listeners {
		server.AddListener(lis)
	}
	server.Register("login", func(ctx *toyframe.Context) error {
		return loginHandler(ctx, t)
	})
	server.Register("send", func(ctx *toyframe.Context) error {
		return sendHandler(ctx, t)
	})
	server.Register("logout", func(ctx *toyframe.Context) error {
		return logoutHandler(ctx, t)
	})

	return server, nil
}

func chatClientSender(reciever string, dial dialer.DialFunc, wg *sync.WaitGroup, t *testing.T) {
	defer wg.Done()
	ctx, err := toyframe.Call("tcp", "localhost:8888", "send", dial,
		&ChatMsg_Send{Reciever: reciever, Content: "hello1"},
		&ChatMsg_Send{Reciever: reciever, Content: "hello2"},
		&ChatMsg_Send{Reciever: reciever, Content: "hello3"},
		&ChatMsg_Send{Reciever: reciever, Content: "hello4"},
		&ChatMsg_Send{Reciever: reciever, Content: "hello5"})
	if err != nil {
		t.Errorf("call send failed: %v", err)
		return
	}
	ctx.Close()
}

func chatClientReciever(name string, dial dialer.DialFunc, wg *sync.WaitGroup, t *testing.T) {
	defer wg.Done()
	ctx, err := toyframe.Call("tcp", "localhost:8888", "login", dial, &ChatMsg_Login{Name: name})
	if err != nil {
		t.Errorf("call login failed: %v", err)
		return
	}
	defer ctx.Close()

	msg_count := 0
	for err == nil {
		tmp := ChatMsg_Send{}
		if err = ctx.ReadObj(&tmp); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("read msg list failed: %v", err)
			return
		}
		msg_count += 1
		if msg_count == 5 {
			go func() {
				ctx_logout, err_logout := toyframe.Call("tcp", "localhost:8888", "logout", dial, &ChatMsg_Login{Name: name})
				if err_logout != nil {
					t.Errorf("call logout failed: %v", err_logout)
					return
				}
				ctx_logout.Close()
			}()
		}
	}
}

func TestChatExample(t *testing.T) {
	toyframe.SetLogWriter(os.Stdout)

	cert, _ := tls.X509KeyPair(tlsPem, tlsKey)

	server_conf := &tls.Config{Certificates: []tls.Certificate{cert}}
	client_conf := &tls.Config{InsecureSkipVerify: true}

	lis, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		t.Errorf("listen failed: %v\n", err)
		return
	}
	defer lis.Close()
	lis = listener.B(lis).WithBitReverse().WithTls(server_conf).WithMultiplex().WithBrotli().Build()
	dial := dialer.B(net.Dial).WithBitReverse().WithTls(client_conf).WithMultiplex().WithBrotli().Build()

	server, _ := setupChatServer(t, lis)
	server_close_ch := make(chan struct{})
	go func() {
		server.Run()
		close(server_close_ch)
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go chatClientReciever("myname", dial, &wg, t)
	go chatClientSender("myname", dial, &wg, t)

	wg.Wait()
	server.Close()
	<-server_close_ch
}
