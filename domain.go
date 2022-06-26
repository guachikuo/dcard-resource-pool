package pool

import (
	"context"
	"fmt"
)

var (
	ErrCreatorNotExist = fmt.Errorf("you must provide a creator")

	ErrDestroyerNotExist = fmt.Errorf("you must provide a destroyer")

	ErrInvalidMaxIdleSize = fmt.Errorf("invalid maxIdleSize")
)

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (T, error)
	// This releases an active resource back to the resource pool
	Release(context.Context, T)
	// This returns the number of idle items
	NumIdle() int
}
