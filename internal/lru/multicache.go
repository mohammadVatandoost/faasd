package lru

const (
	FoCCache  = 0
	TAHCCache = 1
)

type MultiCache struct {
	focCache *Cache
	tahc     *Cache
}

func (mc *MultiCache) Add(cacheSelector int, key Key, value interface{}) {
	if cacheSelector == FoCCache {
		mc.focCache.Add(key, value)
	} else {
		mc.tahc.Add(key, value)
	}
}

func (mc *MultiCache) Get(cacheSelector int, key Key) (value interface{}, ok bool) {
	if cacheSelector == FoCCache {
		return mc.focCache.Get(key)
	} else {
		return mc.tahc.Get(key)
	}
}

func (mc *MultiCache) Resize(focMaxSize int64, tahcMaxSize int64) {
	mc.focCache.Resize(focMaxSize)
	mc.tahc.Resize(tahcMaxSize)
}

func NewMultiCache(focMaxSize int64, tahcMaxSize int64) *MultiCache {
	return &MultiCache{
		focCache: New(focMaxSize),
		tahc:     New(tahcMaxSize),
	}
}
