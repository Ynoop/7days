package ycache

import (
	"context"
	"errors"
	"log"
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

// Group 每个group都是cache的命名空间，并加载相关数据
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
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
		log.Println("[ycache] mainCache.get hit")
		return value, nil
	}

	// 本地加载数据
	return g.load(ctx, key)
}

// RegisterPeers 注册一个远程选择的分布式节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}

	g.peers = peers
}

// 加载数据
func (g *Group) load(ctx context.Context, key string) (value ByteView, err error) {
	if g.peers != nil {
		// 如果有远程节点。从远程节点中加载数据
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(ctx, peer, key); err == nil {
				return value, nil
			}

			log.Println("[YCache] Failed to get from peer", err)
		}
	}

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

func (g *Group) getFromPeer(ctx context.Context, peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(ctx, g.name, key)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{data: bytes}, nil
}
