package pool

import (
	"context"
	"sync"
	"time"
)

type resource[T any] struct {
	value T
	// the time when this resource is created
	createdAt time.Time
}

type pool[T any] struct {
	mutex sync.RWMutex

	resource chan resource[T]

	creator   func(context.Context) (T, error)
	destroyer func(context.Context, T)

	maxIdleTime time.Duration
}

func New[T any](
	creator func(context.Context) (T, error),
	destroyer func(context.Context, T),
	maxIdleSize int,
	maxIdleTime time.Duration,
) (Pool[T], error) {
	ctx := context.Background()

	if creator == nil {
		return nil, ErrCreatorNotExist
	}

	if destroyer == nil {
		return nil, ErrDestroyerNotExist
	}

	if maxIdleSize <= 0 {
		return nil, ErrInvalidMaxIdleSize
	}

	// init
	p := &pool[T]{
		resource: make(chan resource[T], maxIdleSize),

		creator:   creator,
		destroyer: destroyer,

		maxIdleTime: maxIdleTime,
	}

	for i := 0; i < maxIdleSize; i++ {
		r, err := p.create(ctx)
		if err != nil {
			// release created resource
			close(p.resource)
			for r := range p.resource {
				p.destroy(ctx, r.value)
			}
			p.resource = nil

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
	for {
		select {
		case resource := <-p.resource:
			if resource.createdAt.Add(p.maxIdleTime).Before(time.Now()) {
				p.destroy(ctx, resource.value)
				continue
			}
			return resource.value, nil
		default:
			return p.create(ctx)
		}
	}
}

func (p *pool[T]) Release(ctx context.Context, r T) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	resource := resource[T]{
		value:     r,
		createdAt: time.Now(),
	}

	select {
	case p.resource <- resource:
	// when the channel is full, default will be executed,
	// then we need to destroy the resource
	default:
		p.destroy(ctx, r)
	}
}

func (p *pool[T]) NumIdle() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.resource)
}

func (p *pool[T]) create(ctx context.Context) (T, error) {
	return p.creator(ctx)
}

func (p *pool[T]) destroy(ctx context.Context, r T) {
	p.destroyer(ctx, r)
}
