package buffered

import "sync/atomic"

type typedAtomic[T any] struct {
	val atomic.Value
}

func (a *typedAtomic[T]) Load() T {
	return a.val.Load().(T)
}

func (a *typedAtomic[T]) Store(newVal T) {
	a.val.Store(newVal)
}
