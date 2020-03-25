package cache

import (
	"sync"
	"time"
)

type elm struct {
	deadline time.Time
	item     interface{}
}
// Cache : Warning! Don't use this cache's implementation in you production codes.
// This cache is only used in example codes.
type Cache struct {
	store  map[string]elm
	locker sync.Mutex
}

func (c *Cache) init() {
	if c.store == nil {
		c.store = make(map[string]elm)
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.init()
	elm, ok := c.store[key]
	if !ok {
		return nil, false
	}
	if elm.deadline.After(time.Now()) {
		delete(c.store, key)
		return nil, false
	}
	return elm.item, true
}

func (c *Cache) Set(key string, value interface{}, exp time.Duration) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.init()
	c.store[key] = elm{time.Now().Add(exp), value}
}
