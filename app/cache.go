package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var KeyExpired = errors.New("key expired")

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
		fmt.Println("returning key expired error")
		return "", KeyExpired
	}
	return v.val, nil
}

func (c *Cache) Store(k string, v Value) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.v[k] = v

	return nil
}
