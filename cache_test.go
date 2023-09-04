package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewEager(t *testing.T) {
	var timesCalled int

	cache, err := NewEager(func() (string, error) {
		timesCalled++
		return "string", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, timesCalled)

	s, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 1, timesCalled)

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesCalled)

	cache.Invalidate()
	s, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 2, timesCalled)
}

func TestNew(t *testing.T) {
	var timesCalled int

	cache := New(func() (string, error) {
		timesCalled++
		return "string", nil
	})
	assert.Equal(t, 0, timesCalled)

	s, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 1, timesCalled)

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesCalled)

	cache.Invalidate()
	s, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 2, timesCalled)
}

func TestCache_SetTTL(t *testing.T) {
	const (
		ttl = 50 * time.Millisecond
	)
	var (
		timesCalled int
	)

	cache := New(func() (string, error) {
		timesCalled++
		return "string", nil
	})
	assert.Equal(t, 0, timesCalled)
	cache.SetTTL(ttl)

	s, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 1, timesCalled)

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesCalled)
	time.Sleep(2 * ttl)

	s, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 2, timesCalled)
}

func TestNewWithTTL(t *testing.T) {
	const (
		ttl = 50 * time.Millisecond
	)
	var (
		timesCalled int
	)

	cache := NewWithTTL(func() (string, time.Duration, error) {
		timesCalled++
		return "string", ttl, nil
	})
	assert.Equal(t, 0, timesCalled)

	s, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 1, timesCalled)

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesCalled)
	time.Sleep(2 * ttl)

	s, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 2, timesCalled)
}

func TestNewWithTTL_Override(t *testing.T) {
	const (
		ttl = 50 * time.Millisecond
	)
	var (
		timesCalled int
	)

	cache := NewWithTTL(func() (string, time.Duration, error) {
		timesCalled++
		return "string", ttl, nil
	})
	assert.Equal(t, 0, timesCalled)
	cache.SetTTL(time.Minute)

	s, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 1, timesCalled)

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesCalled)
	time.Sleep(2 * ttl)

	s, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, "string", s)
	assert.Equal(t, 2, timesCalled, "1 minute TTL should have been overridden by loader")
}
