package lru

import (
	"fmt"
	"github.com/xiaobinqt/cat/lru/pb"
	"github.com/xiaobinqt/cat/lru/singleflight"
	"log"
	"sync"
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker

	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}

	if _, ok := groups[name]; ok {
		panic(fmt.Sprintf("name : %s existed", name))
	}

	groups[name] = g
	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}

	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{
		b: res.Value,
	}, nil
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("cache key [%s] hit", key)
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.

	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[catCache] Failed to get from peer:", err.Error())
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return ByteView{}, err
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: bytes}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
