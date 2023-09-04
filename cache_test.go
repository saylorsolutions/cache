package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestValue(t *testing.T) {
	const expected = "a string"
	cache := Value(expected)
	val, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, expected, val)

	cache.Invalidate()
	val, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, expected, val)
}
