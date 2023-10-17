package buffered

import "sync/atomic"

type TypedAtomic[T any] struct {
	val atomic.Value
}

func (a *TypedAtomic[T]) Load() T {
	return a.val.Load().(T)
}

func (a *TypedAtomic[T]) Store(newVal T) {
	a.val.Store(newVal)
}
