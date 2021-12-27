package ycache

import (
	"context"
	"errors"
	"sync"
)

// Getter 从key中加载数据
// 作为Group未命中数据时的回调函数
type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// GetterFunc 用以实现Getter的方法
type GetterFunc func(ctx context.Context, key string) ([]byte, error)

func (f GetterFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return f(ctx, key)
}

// Group 每个group都是cache的命名空间，允许加载的数据存放在各处
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 初始化Group
func NewGroup(name string, cacheBytes int, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = g

	return g
}

// 返回一个group实例，如果groups中没有改值，则返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(ctx context.Context, key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is requeired")
	}

	// 缓存中获取
	if value, hit := g.mainCache.get(key); hit {
		return value, nil
	}

	// 本地加载数据
	return g.load(ctx, key)
}

// 加载数据
func (g *Group) load(ctx context.Context, key string) (ByteView, error) {
	return g.getLocally(ctx, key)
}

// 从回调函数中加入数据
func (g *Group) getLocally(ctx context.Context, key string) (ByteView, error) {
	bytes, err := g.getter.Get(ctx, key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{data: cloneBytes(bytes)}
	g.populateCache(key, value)

	return value, nil
}

// 把数据插入至缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
