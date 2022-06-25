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
	mutex     sync.RWMutex
	closeOnce sync.Once

	resource chan resource[T]

	creator func(context.Context) (T, error)
	closer  func(context.Context, T)
}

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (T, error)
	// This releases an active resource back to the resource pool
	Release(context.Context, T)
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

		creator: creator,
		closer:  closer,
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

func (p *pool[T]) Release(ctx context.Context, r T) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	resource := resource[T]{
		value:     r,
		createdAt: time.Now(),
	}

	select {
	case p.resource <- resource:
	default:
		p.closer(ctx, r)
	}
}

func (p *pool[T]) NumIdle() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.resource)
}

func (p *pool[T]) close(ctx context.Context) {
	p.closeOnce.Do(
		func() {
			close(p.resource)
			for r := range p.resource {
				p.closer(ctx, r.value)
			}

			p.resource = nil
		},
	)
}
