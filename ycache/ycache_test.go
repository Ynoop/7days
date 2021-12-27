package ycache

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")

	if v, _ := f.Get(context.Background(), "key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("expect: %s, got %s", expect, v)
	}
}

func TestGet(t *testing.T) {
	db := map[string]string{
		"zhangsan": "fwkt",
		"lisi":     "lhsm",
		"zhaowu":   "luren",
	}

	loadCount := make(map[string]int, len(db))
	g := NewGroup("peple", 2<<10, GetterFunc(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Println("[SlowDB] search key:", key)
			if v, ok := db[key]; ok {
				if _, hit := loadCount[key]; hit {
					loadCount[key] = 0
				}

				loadCount[key] += 1
				return []byte(v), nil
			}

			return nil, fmt.Errorf("%s not exist", key)
		}))

	ctx := context.Background()
	for k, v := range db {
		if view, err := g.Get(ctx, k); err != nil || view.String() != v {
			t.Fatalf("want get %s, got %s", v, view.String())
		}
		if _, err := g.Get(ctx, k); err != nil || loadCount[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := g.Get(ctx, "unknow"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
