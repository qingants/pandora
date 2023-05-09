package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/qingants/pandora/minirpc"
)

type Foo int

type Args struct {
	A int
	B int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func startServer(addr chan string) {
	var foo Foo
	if err := minirpc.Register(&foo); err != nil {
		log.Fatal("register error ", err)
	}
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error ", err)
	}
	log.Printf("start rpc server on %s", l.Addr())
	addr <- l.Addr().String()
	minirpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	// l, err := net.Listen("tcp", "127.0.0.1:7979")
	// if err != nil {
	// 	log.Fatalf("network error %s", err.Error())
	// }
	// log.Printf("rpc server server on: %s", l.Addr())
	// minirpc.Accept(l)
	addr := make(chan string)
	go startServer(addr)
	client, _ := minirpc.Dial("tcp", <-addr)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second * 2)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{
				A: i,
				B: i * i,
			}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error ", err)
			}
			log.Printf("%d + %d = %d", args.A, args.B, reply)
		}(i)
	}
	wg.Wait()
}
