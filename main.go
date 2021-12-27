package main

import (
	"7days/ycache"
	"context"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Zhangsan": "1",
	"Lisi":     "2",
	"Wangwu":   "3",
}

func main() {
	ycache.NewGroup("names", 2<<10, ycache.GetterFunc(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}

			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := ycache.NewHTTPPool(addr)
	log.Println("ycache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
