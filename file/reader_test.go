package file

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type _testData struct {
	Name string `json:"name"`
	Desc string `json:"description"`
}

func TestNewReaderCache(t *testing.T) {
	var (
		timesFetched     int
		timesInvalidated int
	)

	tmp, err := os.MkdirTemp("", "NewReaderCache-*")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()
	filename := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(filename, []byte(`{"name":"Go","description":"A super cool language and ecosystem"}`), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache, err := NewReaderCache[_testData](ctx, filename, func(reader io.Reader) (_testData, error) {
		timesFetched++
		var data _testData
		err := json.NewDecoder(reader).Decode(&data)
		return data, err
	}, testingLog(t))
	assert.NoError(t, err)
	assert.Equal(t, 0, timesFetched, "Should not have fetched yet, since the reader cache is lazy initialized")
	cache.OnInvalidate(func() {
		timesInvalidated++
	})

	data, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, _testData{Name: "Go", Desc: "A super cool language and ecosystem"}, data)
	assert.Equal(t, 1, timesFetched, "Should have fetched once")
	assert.Equal(t, 0, timesInvalidated, "Cache should not have been invalidated yet")

	_, _ = cache.Get()
	_, _ = cache.Get()
	_, _ = cache.Get()
	assert.Equal(t, 1, timesFetched, "Data should still be cached at this point")
	assert.Equal(t, 0, timesInvalidated, "Cache should not have been invalidated yet")
	cache.Invalidate()
	assert.Equal(t, 1, timesInvalidated, "Invalidate func should have been called now")

	data, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, _testData{Name: "Go", Desc: "A super cool language and ecosystem"}, data)
	assert.Equal(t, 2, timesFetched, "Invalidation should have resulted in another fetch with the same data")
	assert.Equal(t, 1, timesInvalidated, "No further Invalidate calls should have happened")
}
