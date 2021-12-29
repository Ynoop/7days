package singleflight

import "sync"

// call 表示正在进行中或者已经结束的请求。
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group singleflight的主数据结构，管理不同key的请求(call)
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		// 如果有正在请求的项目，则等待
		g.mu.Unlock()
		c.wg.Wait()
		// 请求结束返回结果
		return c.val, c.err
	}

	c := new(call)
	// 发起请求加锁
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 调用函数发起请求
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	// 跟新g.m
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
