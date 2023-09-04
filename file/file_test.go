package file

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEagerFileCache(t *testing.T) {
	tmp, err := os.MkdirTemp("", "NewEagerFileCache-*")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()
	filename := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(filename, []byte("Hello!"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache, err := NewEagerFileCache(ctx, filename, testingLog(t))
	assert.NoError(t, err)

	data, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello!"), data)

	assert.NoError(t, os.WriteFile(filename, []byte("Another message"), 0644))
	time.Sleep(100 * time.Millisecond)
	data, err = cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Another message"), data)
}
