// LRU 最近最少使用算法
// 本项目受到geektutu大神的Blog启发，自己模仿groupcache项目中LRU算法
// groupcache:https://github.com/golang/groupcache
//
// LRU 算法本质是使用一个hash table 维护缓存内容，
//     使用一个列队维护缓存使用的频率。
//     当缓存使用数到达最大时，将弹出队首元素，并删除hash table 中的缓存内容。
// 这里把表尾作为最近最少使用，表首作为最近经常使用
package lru

import (
	"container/list"
)

type Cache struct {
	// 最大实例个数
	MaxEntries int

	// 双向链表用以存储每个缓存最近使用频率，表尾为最近最少使用，表首为最近最多使用
	ll *list.List

	cache map[interface{}]*list.Element

	// 回调函数，当缓存被销魂是执行
	OnEvicted func(key Key, value interface{})
}

type Key interface{}

// 缓存实体
type entry struct {
	key   Key
	value interface{}
}

// 创建一个新的缓存
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// 增加新元素
func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil {
		// 如果缓存为空的情况
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	// 判断当前key是否存在
	if ee, ok := c.cache[key]; ok {
		// 把当前元素移动至表首
		c.ll.MoveToFront(ee)
		// 修改缓存内容
		ee.Value.(*entry).value = value
		return
	}

	// 把元素加入到表首
	ele := c.ll.PushBack(&entry{key, value})
	c.cache[key] = ele

	// 判断是否到达缓存最大值
	if c.MaxEntries != 0 && c.MaxEntries < c.ll.Len() {
		c.RemoveOldest()
	}
}

// 删除最少使用的元素
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}

	// 找到表尾的第一个元素
	ele := c.ll.Back()
	c.removeElement(ele)
}

// 根据key查找指定元素
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}

	if ele, hit := c.cache[key]; hit {
		// 把当前元素移动至表首
		c.ll.MoveToFront(ele)

		return ele.Value.(*entry).value, true
	}

	return
}

// 根据指定key移除缓存元素
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}

	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// 删除元素
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)

	// 执行回调
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// 返回当前缓存个数
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}

	return c.ll.Len()
}
