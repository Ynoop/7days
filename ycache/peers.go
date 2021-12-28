// 分布式节点，主要用以声明节点的接口
package ycache

import "context"

//  PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(ctx context.Context, group string, key string) ([]byte, error)
}
