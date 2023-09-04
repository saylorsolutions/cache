package file

import (
	"context"
	"github.com/saylorsolutions/cache"
	"io"
)

// NewFileCache creates a new [cache.Value] that reads the given file in its loader.
func NewFileCache(ctx context.Context, filename string, log NotifyLog) (*cache.Value[[]byte], error) {
	return NewReaderCache[[]byte](ctx, filename, io.ReadAll, log)
}

// NewEagerFileCache is the same as NewFileCache, except that it will proactively read the file's contents into memory.
func NewEagerFileCache(ctx context.Context, filename string, log NotifyLog) (*cache.Value[[]byte], error) {
	fileCache, err := NewFileCache(ctx, filename, log)
	if err != nil {
		return nil, err
	}
	_, err = fileCache.Get()
	if err != nil {
		return nil, err
	}
	return fileCache, nil
}
