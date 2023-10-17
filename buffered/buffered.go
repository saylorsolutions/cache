package buffered

import (
	"fmt"
	"github.com/saylorsolutions/cache"
	"time"
)

type MultiCache[K comparable, V any] struct {
	readCache   *cache.MultiCache[K, V]
	writeBuffer *cache.MultiCache[K, *TypedAtomic[V]]
}

// NewMulti will create a new MultiCache.
// A MultiCache may be composed of other MultiCache in the case where logical grouping of cached values is needed.
func NewMulti[K comparable, V any]() *MultiCache[K, V] {
	buffer := cache.NewMulti[K, *TypedAtomic[V]](func(key K) (*TypedAtomic[V], error) {
		val := new(TypedAtomic[V])
		var mt V
		val.Store(mt)
		return val, nil
	})
	reader := cache.NewMulti[K, V](func(key K) (V, error) {
		atom, err := buffer.Get(key)
		if err != nil {
			var mt V
			return mt, err
		}
		return atom.Load(), nil
	})
	return &MultiCache[K, V]{
		readCache:   reader,
		writeBuffer: buffer,
	}
}

// Preheat will load the values associated with each key in keys into the read cache.
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

// Get will return the value in the read cache associated with key K.
// Any errors returned from [cache.Value.Get] will be returned from Get.
func (m *MultiCache[K, V]) Get(key K) (V, error) {
	return m.readCache.Get(key)
}

// MustGet does the same thing as Get, but it will panic if an error occurs.
func (m *MultiCache[K, V]) MustGet(key K) V {
	val, err := m.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Invalidate will invalidate the cache.Value related to key K, if it exists.
func (m *MultiCache[K, V]) Invalidate(key K) {
	m.readCache.Invalidate(key)
}

// OnInvalidate sets a [cache.OnInvalidateFunc] on the Value referenced by key.
// If no Value is associated to the given key, then no action is taken.
func (m *MultiCache[K, V]) OnInvalidate(key K, invalidateFunc cache.OnInvalidateFunc) {
	m.readCache.OnInvalidate(key, invalidateFunc)
}

// Set will set the key in the write buffer to the assigned value.
// This implicitly invalidates the same key in the read cache.
func (m *MultiCache[K, V]) Set(key K, val V) {
	atom := m.writeBuffer.MustGet(key)
	atom.Store(val)
	m.Invalidate(key)
}

// Unset will clear a value referenced by key in the MultiCache.
// Which means the next Get call for the same key will return the default value for V.
func (m *MultiCache[K, V]) Unset(key K) {
	m.writeBuffer.Invalidate(key)
	m.readCache.Invalidate(key)
}

// SetTTLPolicy sets the time to live policy for all read cache values after they are retrieved.
// By default, a MultiCache value will not invalidate itself.
// A TTL policy must be set prior to retrieval or preheating for any value to invalidate itself.
//
// Note that this policy takes precedence over any individual [Value]'s time to live value.
//
// This method will panic if ttl <= 0.
func (m *MultiCache[K, V]) SetTTLPolicy(ttl time.Duration) {
	m.readCache.SetTTLPolicy(ttl)
}
