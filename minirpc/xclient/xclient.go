package xclient

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/qingants/pandora/minirpc"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *minirpc.Option
	lock    sync.Mutex
	clients map[string]*minirpc.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *minirpc.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*minirpc.Client),
	}
}

func (xc *XClient) Close() error {
	xc.lock.Lock()
	defer xc.lock.Unlock()

	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*minirpc.Client, error) {
	xc.lock.Lock()
	defer xc.lock.Unlock()
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}

	if client == nil {
		var err error
		client, err = minirpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (cx *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply any) error {
	client, err := cx.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	// log.Printf("-----call addr--- %s\n", rpcAddr)
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// Broadcast invokes the named function for every server register in discovery
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply any) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)

	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var cloneReply any
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, cloneReply)
			lock.Lock()
			defer lock.Unlock()

			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}
		}(rpcAddr)
	}

	wg.Wait()
	return e
}
