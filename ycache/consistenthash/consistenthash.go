package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash map 用bytes key 获取 uint32的hash值
type Hash func(data []byte) uint32

// Map 包含了所有的hash keys
type Map struct {
	hash     Hash
	replicas int   // 虚拟节点数
	keys     []int // Sorted
	hashMap  map[int]string
}

// New create Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		// 默认的hash算法
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add 将Key添加至hashMap
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 创建对应的虚拟节点
			hash := int(m.hash([]byte(strconv.FormatInt(int64(i), 10) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}

	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
