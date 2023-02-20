package minirpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Call struct {
	Seq    uint64
	Method string
	Args   any
	Reply  any
	Error  error
	Done   chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	codec    Codec
	opt      *Option
	head     Head
	lock     sync.Mutex
	sending  sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

var _ io.Closer = (*Client)(nil)
var ErrShutdown = errors.New("connection is shutdown")

func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.codec.Close()
}

func (c *Client) IsAvailable() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return !c.closing && !c.shutdown
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}

	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.lock.Lock()
	defer c.lock.Unlock()
	// if c.closing || c.shutdown {
	// 	return nil
	// }
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()

	c.lock.Lock()
	defer c.lock.Unlock()

	c.shutdown = true
	for _, call := range c.pending {
		call.Error = nil
		call.done()
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		var h Head
		if err = c.codec.ReadHead(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.codec.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = c.codec.ReadBody(nil)
			call.done()
		default:
			err = c.codec.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body error " + err.Error())
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, errors.New("incorrect number of options")
	}

	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client failed to create codec:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client failed to encode options:", err)
		_ = conn.Close()
		return nil, err
	}

	return NewClientCodec(f(conn), opt), nil
}

func NewClientCodec(codec Codec, opt *Option) *Client {
	client := &Client{
		seq:     1,
		codec:   codec,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func Dial(network, addr string, opts ...*Option) (*Client, error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

// func DialTimeout(network, addr string, timeout time.Duration, opts *Option) (*Client, error) {
// 	conn, err := net.DialTimeout(network, addr, timeout)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewClient(conn, opts)
// }

func (c *Client) Send(call *Call) {
	log.Println("rpc client sending request:", call.Seq)

	c.sending.Lock()
	defer c.sending.Unlock()

	// log.Println("write...1.")
	seq, err := c.registerCall(call)
	if err != nil {
		log.Printf("rpc client register call: %v", err)
		call.Error = err
		call.done()
		return
	}
	// log.Println("write...2.")
	c.head.Method = call.Method
	c.head.Seq = seq
	c.head.Error = ""

	// log.Println("write...3.")
	if err := c.codec.Write(&c.head, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) Call(method string, argv, reply any) error {
	call := <-c.Go(method, argv, reply, make(chan *Call, 1)).Done
	return call.Error
}

func (c *Client) Go(method string, argv, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		Method: method,
		Args:   argv,
		Reply:  reply,
		Done:   done,
	}
	c.Send(call)
	return call
}
