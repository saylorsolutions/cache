package file

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/saylorsolutions/cache"
	"io"
	"os"
	"path/filepath"
)

// NewReaderCache returns a cache of a type extracted from the watched file.
// Whatever type is produced from readFunc will be the type of the [cache.Value], which makes this useful for unmarshalling a file's contents into a user defined type.
func NewReaderCache[T any](ctx context.Context, filename string, readFunc func(io.Reader) (T, error), log NotifyLog) (*cache.Value[T], error) {
	orig := filename
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for '%s': %w", orig, err)
	}
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to stat file '%s': %w", filename, err)
	}
	if fi.IsDir() {
		return nil, errors.New("unable to cache directories")
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	loader := cache.LoaderFunc[T](func() (T, error) {
		var t T
		f, err := os.Open(filename)
		if err != nil {
			cancel()
			return t, fmt.Errorf("failed to open file '%s' for reading: %w", filename, err)
		}
		defer func() {
			_ = f.Close()
		}()

		t, err = readFunc(f)
		if err != nil {
			cancel()
			return t, fmt.Errorf("failed to read file '%s' contents: %w", filename, err)
		}
		return t, nil
	})
	_cache := cache.New(loader)

	watchDir := filepath.Dir(filename)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to create filesystem watcher: %w", err)
	}

	if log == nil {
		log = newNoOpNotifyLog()
	}

	go func() {
		defer func() {
			err := watcher.Close()
			if err != nil {
				log.Error(fmt.Errorf("failed to close watcher for goroutine exit: %w", err))
			}
		}()
		err := watcher.Add(watchDir)
		if err != nil {
			log.Error(fmt.Errorf("failed to add file '%s' to watcher: %w", filename, err))
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-watcher.Events:
				if evt.Name != filename {
					log.UnrelatedEvent(evt)
					continue
				}
				log.Event(evt)
				// Any change could indicate the need for a reload.
				// The loader will cancel the context if there's a hard stop error, so we don't need to handle the various Op cases here.
				_cache.Invalidate()
			case err := <-watcher.Errors:
				log.Error(fmt.Errorf("error watching cache file '%s': %w", filename, err))
			}
		}
	}()
	return _cache, nil
}
