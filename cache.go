package cache

import (
	"errors"
	"sync"
	"time"
)

// LoaderFunc is a function that loads data for type T.
type LoaderFunc[T any] func() (T, error)

// BreakerFunc is a function that breaks the cache by setting a pointer to type T to nil.
type BreakerFunc[T any] func()

// Value is a struct that contains a value of type T.
// If the value is nil, then its LoaderFunc will be called.
type Value[T any] struct {
	val      *T
	loadFunc LoaderFunc[T]
	lock     sync.RWMutex
	ttl      time.Duration
}

// New creates a new, lazily initialized Value with the given loader.
// If the loader is nil, then this function will panic.
func New[T any](loader LoaderFunc[T]) *Value[T] {
	if loader == nil {
		panic("nil loader")
	}
	return &Value[T]{
		loadFunc: loader,
	}
}

// NewEager will create an eagerly initialized Value with the given loader.
// If the loader is nil, then this function will panic.
func NewEager[T any](loader LoaderFunc[T]) (*Value[T], error) {
	cache := New(loader)
	_, err := cache.Get()
	if err != nil {
		return nil, err
	}
	return cache, nil
}

// Get will return the cached value, if it exists, or call the LoaderFunc otherwise.
// Any error returned while loading the cache will be returned.
func (c *Value[T]) Get() (T, error) {
	c.lock.RLock()
	if c.val != nil {
		val := *c.val
		c.lock.RUnlock()
		return val, nil
	}
	c.lock.RUnlock()
	return c.load()
}

func (c *Value[T]) load() (T, error) {
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
	if c.ttl > 0 {
		time.AfterFunc(c.ttl, c.Invalidate)
	}
	return val, nil
}

// Invalidate will remove the cached value and force a reload the next time Get is called.
func (c *Value[T]) Invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.val = nil
}

// SetTTL sets the duration that a cached value will persist after a call to Get.
// Note that using SetTTL with NewEager will have no effect because the value has already been cached by the time the Value is created and SetTTL is called.
//
// This is useful for cases where the cached value is expected to change frequently enough that it's only expected to be valid for a short time.
// By default, a Value will not invalidate itself.
// This method must be called to enable that behavior.
//
// If a time to live is no longer needed, then create a new Value with the existing value.
// This function will panic if ttl is <= 0.
func (c *Value[T]) SetTTL(ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ttl <= 0 {
		panic("ttl <= 0")
	}
	c.ttl = ttl
}

// LoaderTTLFunc is a function that returns both a value and the time that the value should be valid.
// If a LoaderTTLFunc returns a time to live <= 0, then an error will be returned from [Value.Get] indicating this.
type LoaderTTLFunc[T any] func() (T, time.Duration, error)

// NewWithTTL will create a Value where the loader determines its own time to live.
// This is useful for cases similar to when the Value holds an authentication token or some other time-valid value, and its time to live is only known upon retrieval.
func NewWithTTL[T any](loader LoaderTTLFunc[T]) *Value[T] {
	if loader == nil {
		panic("nil loader")
	}
	c := new(Value[T])
	c.loadFunc = func() (T, error) {
		// This will run in a locked context, so it's fine to set the ttl.
		val, ttl, err := loader()
		var mt T
		if err != nil {
			return mt, err
		}
		if ttl <= 0 {
			return mt, errors.New("ttl <= 0")
		}
		c.ttl = ttl
		return val, nil
	}
	return c
}
