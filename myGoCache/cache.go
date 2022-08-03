package myGoCache

import (
	"myGoCache/lru_cache"
	"sync"
)

/*
	cache.go
	实例化lru,封装get和add方法,并添加互斥锁mutex实现并发安全
*/
type cache struct {
	mutex      sync.Mutex       //并发互斥锁
	lru        *lru_cache.Cache //实现LRU策略的Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = lru_cache.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
