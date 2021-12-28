package main

import (
	"7days/ycache"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Zhangsan": "1-1",
	"Lisi":     "1-2",
	"Wangwu":   "1-3",
}

func createGroup() *ycache.Group {
	return ycache.NewGroup("names", 2<<10, ycache.GetterFunc(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}

			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, y *ycache.Group) {
	peers := ycache.NewHTTPPool(addr)
	peers.Set(addrs...)
	y.RegisterPeers(peers)
	log.Println("ycache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, y *ycache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := y.Get(r.Context(), key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Context-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "YCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string

	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	y := createGroup()
	if api {
		go startAPIServer(apiAddr, y)
	}

	startCacheServer(addrMap[port], addrs, y)
}
