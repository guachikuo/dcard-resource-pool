package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type resource[T any] struct {
	value T
	// the time when this resource is created
	createdAt time.Time
}

type pool[T any] struct {
	mutex    sync.RWMutex
	resource chan resource[T]
	closer   func(context.Context, T)
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
	closer func(context.Context, T),
	maxIdleSize int,
	maxIdleTime time.Duration,
) (Pool[T], error) {
	ctx := context.Background()

	// init
	p := &pool[T]{
		resource: make(chan resource[T], maxIdleSize),
	}

	for i := 0; i < maxIdleSize; i++ {
		r, err := creator(ctx)
		if err != nil {
			p.close(ctx)
			return nil, err
		}

		p.resource <- resource[T]{
			value:     r,
			createdAt: time.Now(),
		}
	}

	return p, nil
}

func (p *pool[T]) Acquire(ctx context.Context) (T, error) {
	return *new(T), fmt.Errorf("not implemented yet")
}

func (p *pool[T]) Release(T) {
	return
}

func (p *pool[T]) NumIdle() int {
	p.mutex.RLock()
	num := len(p.resource)
	p.mutex.RUnlock()
	return num
}

func (p *pool[T]) close(ctx context.Context) {
	p.mutex.Lock()
	resource := p.resource
	p.resource = nil
	p.mutex.Unlock()

	if resource == nil {
		return
	}

	close(resource)
	for r := range resource {
		p.closer(ctx, r.value)
	}
	return
}
