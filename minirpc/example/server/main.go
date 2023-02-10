package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/qingants/pandora/minirpc"
)

func run(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("network error %s", err.Error())
	}
	log.Printf("rpc server server on: %s", l.Addr())
	addr <- l.Addr().String()
	minirpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go run(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer func() {
		conn.Close()
	}()
	time.Sleep(time.Second * 2)
	_ = json.NewEncoder(conn).Encode(minirpc.DefaultOption)
	cc := minirpc.NewGobCodec(conn)
	for i := 0; i < 10; i++ {
		h := &minirpc.Head{
			Method: "math.Sum",
			Seq:    uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("minirpc req %d", h.Seq))
		_ = cc.ReadHead(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply: ", reply)
	}
}
