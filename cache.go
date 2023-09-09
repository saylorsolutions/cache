package cache

import (
	"errors"
	"sync"
	"time"
)

// LoaderFunc is a function that loads data of type T.
type LoaderFunc[T any] func() (T, error)

// OnInvalidateFunc is a function that will be called when Invalidate is called.
// If no OnInvalidateFunc is set in a Value, then no action will be taken.
type OnInvalidateFunc = func()

// Value is a struct that contains a value of type T.
// If the value is nil, then its LoaderFunc will be called.
// By default, a cached Value will not invalidate itself.
// A time to live must be set to enable this behavior.
//
// See SetTTL for more details.
type Value[T any] struct {
	val      *T
	loadFunc LoaderFunc[T]

	mux          sync.RWMutex
	ttl          time.Duration
	expiration   time.Time
	getRefreshes bool
	onInvalidate OnInvalidateFunc
}

func (c *Value[T]) cacheExpired() bool {
	if c.ttl <= 0 {
		return false
	}
	return c.expiration.Before(time.Now())
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
	c.mux.RLock()
	if c.val != nil && !c.cacheExpired() {
		val := *c.val
		getRefreshes := c.getRefreshes
		c.mux.RUnlock()
		if getRefreshes {
			c.refreshTimer()
		}
		return val, nil
	}
	c.mux.RUnlock()
	return c.load()
}

func (c *Value[T]) refreshTimer() {
	c.mux.RLock()
	ttl := c.ttl
	c.mux.RUnlock()
	if ttl <= 0 {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.expiration = time.Now().Add(ttl)
}

func (c *Value[T]) load() (T, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.val != nil && !c.cacheExpired() {
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
		c.expiration = time.Now().Add(c.ttl)
	}
	return val, nil
}

// Invalidate will remove the cached value and force a reload the next time Get is called.
func (c *Value[T]) Invalidate() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.expiration = time.Time{}
	c.val = nil
	if c.onInvalidate != nil {
		c.onInvalidate()
	}
}

// OnInvalidate allows reacting to Invalidate being called on a Value.
// This can be useful in cases where a Value's validity is considered an event where some component of an application needs to be reinitialized.
// This pairs well with a context.CancelFunc.
//
// Note that the given function is only called when Invalidate is called, not when a Value's expiration has been reached.
func (c *Value[T]) OnInvalidate(fn OnInvalidateFunc) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.onInvalidate = fn
}

// SetTTL sets the duration that a cached value will be considered valid.
// This is useful for cases where the cached value is expected to change frequently enough that it's only expected to be valid for a short time.
// This works with NewEager because it sets a static time.Time when the value will be considered invalid.
//
// This method will not allow a call to Get to refresh the expiration time.
// If that's not desired, then use EnableGetTTLRefresh.
//
// If a time to live is no longer needed, then use RemoveTTL.
// This function will panic if ttl is <= 0.
func (c *Value[T]) SetTTL(ttl time.Duration) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if ttl <= 0 {
		panic("ttl <= 0")
	}
	c.ttl = ttl
	c.expiration = time.Now().Add(ttl)
	c.getRefreshes = false
}

// RemoveTTL will remove the time to live constraint on this cached Value.
// This will not invalidate the cache, but Invalidate can be used for that purpose.
func (c *Value[T]) RemoveTTL() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.ttl = 0
	c.expiration = time.Time{}
}

// EnableGetTTLRefresh will change the Get behavior to refresh validity of a cached Value when it's called.
// This is useful for situations where loading should be avoided, and the Value is expected to still be valid after its initial time to live has expired.
//
// This function has no effect if there isn't a time to live set.
func (c *Value[T]) EnableGetTTLRefresh() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.getRefreshes = true
}

// LoaderTTLFunc is a function that returns both a value and the time that the value should be valid.
// If a LoaderTTLFunc returns a time to live <= 0, then an error will be returned from [Value.Get] indicating this.
type LoaderTTLFunc[T any] func() (T, time.Duration, error)

// NewWithTTL will create a Value where the loader determines its value's time to live.
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
