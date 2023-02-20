package main

import (
	"errors"
	"sync"
	"time"
)

type Value struct {
	val string
	exp int64
}

type Cache struct {
	v map[string]Value
	m sync.RWMutex
}

func (c *Cache) Load(k string) (string, error) {
	c.m.RLock()
	defer c.m.RUnlock()

	v, ok := c.v[k]
	if !ok {
		return "", errors.New("key not found")
	}
	if v.exp > 0 && time.Now().UnixMilli() > v.exp {
		delete(c.v, k)
		return "", errors.New("key expired")
	}
	return v.val, nil
}

func (c *Cache) Store(k string, v Value) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.v[k] = v

	return nil
}
