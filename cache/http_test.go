package cache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	go func() {
		log.Println("groupcache is running at", addr)
		log.Fatal(http.ListenAndServe(addr, peers))
	}()
	res, err := http.Get("http://localhost:9999/_groupcache/scores/kkk")
	if err != nil {
		t.Errorf("bug...")
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	fmt.Println(string(body))
	res, err = http.Get("http://localhost:9999/_groupcache/scores/Tom")
	if err != nil {
		t.Errorf("bug...")
	}
	body, err = io.ReadAll(res.Body)
	res.Body.Close()
	fmt.Println(string(body))
}
