package lru

import (
	"container/list"
	"fmt"
	"sync"
)

// Cache is an LRU cache. It is safe for concurrent access.
type Cache struct {
	// MaxSize is the maximum number of all entries size to byte
	MaxSize int64
	// total size of all entries to byte
	Size int64
	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
	mutex sync.Mutex
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   string
	value interface{}
	size  int
}

// Add adds a value to the cache.
func (c *Cache) AddByteArray(key string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	s := len(value) + len(key)

	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		c.Size = c.Size + int64(s)
		ee.Value.(*entry).value = value
		ee.Value.(*entry).size = s
		return
	}
	ele := c.ll.PushFront(&entry{key, value, s})
	c.Size = c.Size + int64(s)
	c.cache[key] = ele
	if c.MaxSize != 0 && c.Size > c.MaxSize {
		fmt.Printf("========== Cache is full, Size: %d ===========\n", c.Size)
		for {
			c.removeOldest()
			if c.Size < c.MaxSize {
				break
			}
		}
	}
}

func (c *Cache) AddUint32(key string, value uint32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	s := 4 + len(key)

	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		c.Size = c.Size + int64(s)
		ee.Value.(*entry).value = value
		ee.Value.(*entry).size = s
		return
	}
	ele := c.ll.PushFront(&entry{key, value, s})
	c.Size = c.Size + int64(s)
	c.cache[key] = ele
	if c.MaxSize != 0 && c.Size > c.MaxSize {
		fmt.Printf("========== Cache is full, Size: %d ===========\n", c.Size)
		for {
			c.removeOldest()
			if c.Size < c.MaxSize {
				break
			}
		}
	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// removeOldest removes the oldest item from the cache.
func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.Size = c.Size - int64(kv.size)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

func (c *Cache) Resize(MaxSize int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.MaxSize = MaxSize
	for {
		if c.Size < c.MaxSize {
			return
		}
		c.removeOldest()
	}
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
