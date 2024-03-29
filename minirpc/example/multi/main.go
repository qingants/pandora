package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/qingants/pandora/minirpc"
	"github.com/qingants/pandora/minirpc/xclient"
)

type Foo int

type Args struct {
	A, B int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.A))
	*reply = args.A + args.B
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := minirpc.NewServer()
	_ = server.Register(&foo)
	addrCh <- l.Addr().String()
	server.Accept(l)
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.A, args.B, reply)
	}
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		_ = xc.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{A: i, B: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{A: i, B: i * i})
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{A: i, B: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)

	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}
