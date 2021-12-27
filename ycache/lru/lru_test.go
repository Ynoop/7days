package lru

import (
	"fmt"
	"testing"
)

type simpleStruct struct {
	int
	string
}

type complexStruct struct {
	int
	simpleStruct
}

var testData = []struct {
	name     string
	keyToAdd interface{}
	keyToGet interface{}
	expected bool
}{
	{"string_hit", "key1", "key1", true},
	{"string_miss", "key2", "val2", false},
	{"simpleStruct_hit", simpleStruct{1, "a"}, simpleStruct{1, "a"}, true},
	{"simpleStruct_miss", simpleStruct{2, "b"}, simpleStruct{1, "a"}, false},
	{"complexStruct_hit", complexStruct{3, simpleStruct{3, "c"}}, complexStruct{3, simpleStruct{3, "c"}}, true},
	{"complexStruct_miss", complexStruct{4, simpleStruct{4, "d"}}, complexStruct{5, simpleStruct{5, "e"}}, false},
}

func TestGet(t *testing.T) {
	for _, d := range testData {
		lru := New(0)
		lru.Add(d.keyToAdd, "val")
		val, ok := lru.Get(d.keyToGet)
		if ok != d.expected {
			t.Fatalf("%s: cache hit = %v, want %v", d.name, ok, !ok)
		} else if ok && val != "val" {
			t.Fatalf("%s expected get to return 'val' but go %v", d.name, val)
		}
	}
}

func TestRemove(t *testing.T) {
	lru := New(0)
	lru.Add("key1", "val1")
	if val, ok := lru.Get("key1"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != "val1" {
		t.Fatalf("TestRemove fataled. Expected %s got %v", "val1", val)
	}

	lru.Remove("key1")
	if _, ok := lru.Get("key1"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}

func TestEvict(t *testing.T) {
	evictedKeys := make([]Key, 0)
	onEvictedFun := func(key Key, val interface{}) {
		evictedKeys = append(evictedKeys, key)
	}

	lru := New(10)
	lru.OnEvicted = onEvictedFun

	for i := 0; i < 20; i++ {
		lru.Add(fmt.Sprintf("key%d", i), 123)
	}

	if len(evictedKeys) < 10 {
		t.Fatalf("go %d evicted keys; want 10", len(evictedKeys))
	}

	if evictedKeys[0] != Key("key0") {
		t.Fatalf("go %v in first evicted key; want %s", evictedKeys[0], "key0")
	}
	t.Logf("%+v\n", evictedKeys)
}
