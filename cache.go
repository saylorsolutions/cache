package cache

import (
	"sync"
)

// LoaderFunc is a function that loads data for type T.
type LoaderFunc[T any] func() (T, error)

// BreakerFunc is a function that breaks the cache by setting a pointer to type T to nil.
type BreakerFunc[T any] func()

// Cache is a struct that contains a value of type T.
// If the value is nil, then its LoaderFunc will be called.
type Cache[T any] struct {
	val      *T
	loadFunc LoaderFunc[T]
	lock     sync.RWMutex
}

func New[T any](loader LoaderFunc[T]) *Cache[T] {
	if loader == nil {
		panic("nil loader")
	}
	return &Cache[T]{
		loadFunc: loader,
	}
}

func Value[T any](val T) *Cache[T] {
	return New[T](
		func() (T, error) {
			return val, nil
		},
	)
}

func NewEager[T any](loader LoaderFunc[T]) (*Cache[T], error) {
	cache := New(loader)
	_, err := cache.Get()
	if err != nil {
		return nil, err
	}
	return cache, nil
}

// Get will return the cached value, if it exists, or call the LoaderFunc otherwise.
// Any error returned while loading the cache will be returned.
func (c *Cache[T]) Get() (T, error) {
	c.lock.RLock()
	if c.val != nil {
		val := *c.val
		c.lock.RUnlock()
		return val, nil
	}
	c.lock.RUnlock()
	return c.load()
}

func (c *Cache[T]) load() (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.val != nil {
		return *c.val, nil
	}
	if c.loadFunc == nil {
		panic("nil load func")
	}

	var mt T
	val, err := c.loadFunc()
	if err != nil {
		return mt, err
	}
	c.val = &val
	return val, nil
}

// Invalidate will remove the cached value and force a reload the next time Get is called.
func (c *Cache[T]) Invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.val = nil
}
