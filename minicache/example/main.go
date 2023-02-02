package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/qingants/pandora/minicache"
)

var db = map[string]string{
	"rocky": "luozhihui",
	"amy":   "luozixun",
	"dim":   "luozihan",
	"yoyo":  "zhangyao",
}

func createGroup() *minicache.Group {
	return minicache.NewGroup("scores", 2<<10, minicache.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("[slowDB] search key %s", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exists", key)
	}))
}

func startCacheServer(addr string, addrs []string, mini *minicache.Group) {
	peers := minicache.NewHTTPPool(addr)
	peers.Set(addrs...)
	mini.RegisterPeers(peers)
	log.Println("minicache is running at, ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, mini *minicache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := mini.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))
	log.Println("frontServer is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8001, "minicache server port")
	flag.BoolVar(&api, "api", false, "start a api server")
	flag.Parse()

	apiAddr := "http://127.0.0.1:9999"
	addrMap := map[int]string{
		8001: "http://127.0.0.1:8001",
		8002: "http://127.0.0.1:8002",
		8003: "http://127.0.0.1:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	mini := createGroup()
	if api {
		go startAPIServer(apiAddr, mini)
	}

	startCacheServer(addrMap[port], []string(addrs), mini)
}
