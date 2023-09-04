package lru

import "container/list"

type Cache struct {
	maxBytes int64
	nBytes   int64
	ll       *list.List
	cache    map[string]*list.Element

	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}
