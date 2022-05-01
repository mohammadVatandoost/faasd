package lru

const (
	FoCCache  = 0
	TAHCCache = 1
)

type MultiCache struct {
	focCache *Cache
	tahc     *Cache
}

func (mc *MultiCache) AddUint32(cacheSelector int, key string, value uint32) {
	if cacheSelector == FoCCache {
		mc.focCache.AddUint32(key, value)
	} else {
		mc.tahc.AddUint32(key, value)
	}
}

func (mc *MultiCache) AddByteArray(cacheSelector int, key string, value []byte) {
	if cacheSelector == FoCCache {
		mc.focCache.AddByteArray(key, value)
	} else {
		mc.tahc.AddByteArray(key, value)
	}
}

func (mc *MultiCache) Get(cacheSelector int, key string) (value interface{}, ok bool) {
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
