package lru

import "container/list"

// 实现LRU-K算法
// LRU-K 算法解决“缓存污染问题” 核心思想为命中1次改为命中k次

type LRUKCache struct {
	maxEntires         int                            // 缓存最大上限
	OnEvicted          func(key Key, val interface{}) // 销毁时回调事件
	k                  int                            // 缓存命中的次数
	templateMaxEntires int                            // 临时最大上限
	template           *list.List                     // 临时双向列表
	templateHash       map[interface{}]*list.Element  // 临时缓存hash表
	ll                 *list.List                     // 缓存双向列表
	cache              map[interface{}]*list.Element  // 缓存hash表
}

// 临时数据命中次数
type templateCount struct {
	visited int
	entry
}

func NewLRUKCache(maxEntires int, k int) *LRUKCache {
	return &LRUKCache{
		maxEntires:         maxEntires,
		k:                  k,
		templateMaxEntires: maxEntires,
		template:           list.New(),
		ll:                 list.New(),
		templateHash:       make(map[interface{}]*list.Element),
		cache:              make(map[interface{}]*list.Element),
	}
}

// 新增缓存内容
// 如果缓存内容在临时表中则增加临时表访问次数，表示热点数据
// 如果在缓存表中，则修改缓存内容到队首
func (c *LRUKCache) Add(key Key, value interface{}) {
	if c.cache == nil {
		// 如果缓存为空的情况
		c.ll = list.New()
		c.cache = make(map[interface{}]*list.Element)

		c.template = list.New()
		c.templateHash = make(map[interface{}]*list.Element)
	}

	// 判断当前key是否存在
	if ele, ok := c.cache[key]; ok {
		// 将元素移动到表尾
		c.ll.MoveToBack(ele)
		// 更新缓存内容
		ele.Value.(*entry).value = value

		return
	}

	// 如果缓存不存在的情况，先将缓存放入临时数据中
	c.addToTemplate(key, value)
}

// 加入临时数据中
func (c *LRUKCache) addToTemplate(key Key, value interface{}) {
	var tc *templateCount
	// 判断临时数据是否存在该元素
	if ele, ok := c.templateHash[key]; ok {
		tc = ele.Value.(*templateCount)
		tc.visited++

		// 判断临时数据是否到达K次
		if tc.visited >= c.k {
			// 移除临时数据
			c.removeTemplate(ele)

			// 加入到cache中
			c.addToCache(key, value)
			return
		}

		ele.Value = tc
	} else {
		// 检查数据是否已经存满
		c.templateChecking()
		// 加入到临时数据中
		tc = &templateCount{
			visited: 1,
			entry: entry{
				key:   key,
				value: value,
			},
		}

		ee := c.template.PushBack(tc)
		c.templateHash[key] = ee
	}

	return
}

// 新增元素至缓存中
func (c *LRUKCache) addToCache(key Key, value interface{}) {
	ee := c.ll.PushBack(&entry{key, value})
	c.cache[key] = ee

	// 判断是否到达缓存上限
	if c.ll.Len() > c.maxEntires {
		c.removeOldest()
	}
}

// 移除最少访问数据
func (c *LRUKCache) removeOldest() {
	// 找到表尾的第一个元素
	ele := c.ll.Back()
	c.removeElement(ele)
}

// 移除元素
func (c *LRUKCache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)

	// 执行回调
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// 临时表最大长度检查
func (c *LRUKCache) templateChecking() {
	if c.templateMaxEntires > c.template.Len() {
		ele := c.template.Back()
		c.removeTemplate(ele)
	}
}

// 移除临时表内容
func (c *LRUKCache) removeTemplate(ele *list.Element) {
	// 删除表尾数据
	c.template.Remove(ele)
	kv := ele.Value.(*templateCount)
	delete(c.templateHash, kv.entry.key)
}

// 获取缓存内容
func (c *LRUKCache) Get(key Key) (value interface{}, ok bool) {
	// 先在cache中取
	if ele, hit := c.cache[key]; hit {
		// 移动到表首
		c.ll.MoveToFront(ele)

		return ele.Value.(*entry).value, hit
	}

	// 在template中访问
	if ele, hit := c.templateHash[key]; hit {
		tc := ele.Value.(*templateCount)
		tc.visited++

		// 判断临时数据是否到达K次
		if tc.visited >= c.k {
			// 移除临时数据
			c.removeTemplate(ele)

			// 加入到cache中
			c.addToCache(key, value)
		}

		return ele.Value.(*templateCount).entry.value, hit
	}

	return nil, false
}
