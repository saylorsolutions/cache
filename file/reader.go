package file

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	cacheit "github.com/saylorsolutions/cache"
	"io"
	"os"
	"path/filepath"
)

// NewReaderCache returns a cache of a type extracted from the watched file.
// Whatever type is produced from readFunc will be the type of the [cache.Cache], which makes this useful for unmarshalling a file's contents into a user defined type.
func NewReaderCache[T any](ctx context.Context, filename string, readFunc func(io.Reader) (T, error), log NotifyLog) (*cacheit.Cache[T], error) {
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
	loader := cacheit.LoaderFunc[T](func() (T, error) {
		var t T
		f, err := os.Open(filename)
		if err != nil {
			return t, fmt.Errorf("failed to open file '%s' for reading: %w", filename, err)
		}
		defer func() {
			_ = f.Close()
		}()

		t, err = readFunc(f)
		if err != nil {
			return t, fmt.Errorf("failed to read file '%s' contents: %w", filename, err)
		}
		return t, nil
	})
	cache := cacheit.New(loader)

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
				cache.Invalidate()
			case err := <-watcher.Errors:
				log.Error(fmt.Errorf("error watching cache file '%s': %w", filename, err))
			}
		}
	}()
	return cache, nil
}
