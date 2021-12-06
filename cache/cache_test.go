package cache

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestGetterFunc(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

func TestGroup_Get(t *testing.T) {
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	loadCounts := make(map[string]int, len(db))
	zag := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))

	for k, v := range db {
		if view, err := zag.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := zag.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := zag.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}

func TestGroupGet1(t *testing.T) {
	zag := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[8001] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("[8001] %s not exist", key)
		},
	))

	// 开启api服务
	apiAddr := "http://localhost:9999"
	go startAPIServer(apiAddr, zag)

	startCacheServer(addrMap[8001], addrs, zag)
}

func TestGroupGet2(t *testing.T) {
	zag := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[8002] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("[8002] %s not exist", key)
		},
	))

	startCacheServer(addrMap[8002], addrs, zag)
}

func TestGroupGet3(t *testing.T) {
	zag := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[8003] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("[8003] %s not exist", key)
		},
	))

	startCacheServer(addrMap[8003], addrs, zag)
}

var addrMap = map[int]string{
	8001: "http://localhost:8001",
	8002: "http://localhost:8002",
	8003: "http://localhost:8003",
}

var addrs = []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 开启api服务
func startAPIServer(apiAddr string, zag *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := zag.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))
	log.Println("fronted server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// 开启缓存服务
func startCacheServer(addr string, addrs []string, group *Group) {
	peers := NewHTTPPool(addr)
	peers.Set(addrs...)
	group.RegisterPeers(peers)
	log.Println("cache is running at ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
