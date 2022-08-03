package myGoCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

/*
	测试myGoCache.go, 测试单机并发缓存
*/

// 用一个map模拟本地数据库
var db = map[string]string{
	"Build": "123",
	"Jack":  "589",
	"Joker": "www",
}

// 测试Getter回调接口
func TestGetter(t *testing.T) {

	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatal("callback failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	goCache := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	for k, v := range db {
		if view, err := goCache.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of build")
		}
		if _, err := goCache.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}
	if view, err := goCache.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but %s got", view)
	}
}
