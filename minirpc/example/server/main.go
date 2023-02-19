package main

import (
	"log"
	"net"

	"github.com/qingants/pandora/minirpc"
)

func main() {
	log.SetFlags(0)
	l, err := net.Listen("tcp", "127.0.0.1:7979")
	if err != nil {
		log.Fatalf("network error %s", err.Error())
	}
	log.Printf("rpc server server on: %s", l.Addr())
	minirpc.Accept(l)
}
