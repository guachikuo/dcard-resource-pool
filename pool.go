package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type pool[T any] struct {
	mutex    sync.RWMutex
	resource chan T
}

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (T, error)
	// This releases an active resource back to the resource pool
	Release(T)
	// This returns the number of idle items
	NumIdle() int
}

func New[T any](
	creator func(context.Context) (T, error),
	maxIdleSize int,
	maxIdleTime time.Duration,
) Pool[T] {
	return &pool[T]{
		resource: make(chan T, maxIdleSize),
	}
}

func (p *pool[T]) Acquire(context.Context) (T, error) {
	return *new(T), fmt.Errorf("not implemented yet")
}

func (p *pool[T]) Release(T) {
	return
}

func (p *pool[T]) NumIdle() int {
	return 0
}
