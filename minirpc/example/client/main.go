package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/qingants/pandora/minirpc"
)

func main() {
	log.SetFlags(0)
	addr := "0.0.0.0:7979"

	client, err := minirpc.Dial("tcp", addr)
	if err != nil {
		log.Panicf("dial %s error %s", addr, err)
	}
	defer func() {
		client.Close()
	}()
	time.Sleep(time.Second * 2)
	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("minirpc req %d", i)
			log.Printf("req: %s", args)

			var reply string
			err := client.Call("User.SayHello", args, &reply)
			if err != nil {
				log.Printf("rpc client call SayHello error %s", err.Error())
			} else {
				fmt.Printf("reply %v\n", reply)
			}
		}(i)
	}
	wg.Wait()
}
