package lru

import "container/list"

func New(MaxSize int64) *Cache {
	return &Cache{
		MaxSize:   MaxSize,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}
