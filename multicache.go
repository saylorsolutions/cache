package cache

import (
	"fmt"
	"sync"
	"time"
)

// MultiLoaderFunc is a lot like LoaderFunc, except that it accepts an input key.
type MultiLoaderFunc[K comparable, V any] func(key K) (V, error)

// MultiCache provides the ability to cache multiple values of type V by some comparable key K.
// An example use-case would be caching database entities by primary key.
type MultiCache[K comparable, V any] struct {
	values map[K]*Value[V]
	loader MultiLoaderFunc[K, V]
	lock   sync.RWMutex
	ttl    time.Duration
}

// NewMulti will create a new MultiCache with the given loader.
// A MultiCache may be composed of other MultiCache in the case where logical grouping of cached values is needed.
// If the loader is nil, then this function will panic.
func NewMulti[K comparable, V any](loader MultiLoaderFunc[K, V]) *MultiCache[K, V] {
	if loader == nil {
		panic("nil loader")
	}
	return &MultiCache[K, V]{
		values: map[K]*Value[V]{},
		loader: loader,
	}
}

// Preheat will load the values associated with each key in keys.
// This will return the first error encountered and stop processing further keys.
func (m *MultiCache[K, V]) Preheat(keys []K) error {
	for _, key := range keys {
		_, err := m.Get(key)
		if err != nil {
			return fmt.Errorf("error preheating cache with key '%v': %w", key, err)
		}
	}
	return nil
}

// Get will return the value in the cache.Value associated with key K.
// Any errors returned from [cache.Value.Get] will be returned from Get.
func (m *MultiCache[K, V]) Get(key K) (V, error) {
	m.lock.RLock()
	c, ok := m.values[key]
	if !ok {
		m.lock.RUnlock()
		m.populate(key)
		return m.Get(key)
	}
	defer m.lock.RUnlock()
	return c.Get()
}

// MustGet does the same thing as Get, but it will panic if an error occurs.
func (m *MultiCache[K, V]) MustGet(key K) V {
	val, err := m.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

func (m *MultiCache[K, V]) populate(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.values[key]; ok {
		return
	}

	if m.loader == nil {
		panic("nil loader")
	}
	c := New[V](func() (V, error) {
		key := key
		return m.loader(key)
	})
	if m.ttl > 0 {
		c.SetTTL(m.ttl)
	}
	m.values[key] = c
}

// Invalidate will invalidate the cache.Value related to key K, if it exists.
func (m *MultiCache[K, V]) Invalidate(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	c, ok := m.values[key]
	if !ok {
		return
	}
	c.Invalidate()
	delete(m.values, key)
}

// OnInvalidate sets an OnInvalidateFunc on the Value referenced by key.
// If no Value is associated to the given key, then no action is taken.
func (m *MultiCache[K, V]) OnInvalidate(key K, invalidateFunc OnInvalidateFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	c, ok := m.values[key]
	if !ok {
		return
	}
	c.OnInvalidate(invalidateFunc)
}

// SetTTLPolicy sets the time to live policy for all internal Value values after they are retrieved.
// By default, a MultiCache value will not invalidate itself.
// A TTL policy must be set prior to retrieval or preheating for any value to invalidate itself.
//
// Note that this policy takes precedence over any individual [Value]'s time to live value.
//
// This method will panic if ttl <= 0.
func (m *MultiCache[K, V]) SetTTLPolicy(ttl time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if ttl <= 0 {
		panic("ttl <= 0")
	}
	m.ttl = ttl
}
