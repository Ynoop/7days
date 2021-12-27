package lru

import "testing"

var testData1 = []struct {
	name     string
	keyToAdd interface{}
	keyToGet interface{}
	expected bool
}{
	{"string_hit", "key1", "key1", true},
	{"string_hit", "key1", "key2", false},
}

func TestLRUKGet(t *testing.T) {
	for _, d := range testData1 {
		lruk := NewLRUKCache(2, 2)
		lruk.Add(d.keyToAdd, 123)

		val, ok := lruk.Get(d.keyToGet)
		if ok != d.expected {
			t.Fatalf("%s cache hit = %v, want %v", d.name, ok, !ok)
		} else if ok && val != 123 {
			t.Fatalf("%s expected get to return 123 bu got %v", d.name, val)
		}

		val, ok = lruk.Get(d.keyToGet)
		if ok != d.expected {
			t.Fatalf("%s cache hit = %v, want %v", d.name, ok, !ok)
		} else if ok && val != 123 {
			t.Fatalf("%s expected get to return 123 bu got %v", d.name, val)
		}

		val, ok = lruk.Get(d.keyToGet)
		if ok != d.expected {
			t.Fatalf("%s cache hit = %v, want %v", d.name, ok, !ok)
		} else if ok && val != 123 {
			t.Fatalf("%s expected get to return 123 bu got %v", d.name, val)
		}
	}
}
