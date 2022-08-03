package main

import (
	"fmt"
	"log"
	"myGoCache"
	"myGoCache/http_server"
	"net/http"
)

// 模拟本地database
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	myGoCache.NewGroup("scores", 2<<10, myGoCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := http_server.NewHttpPool(addr)
	log.Println("myGoCache is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
