package pool

import (
	"context"
	"fmt"
)

var (
	ErrCreatorNotExist = fmt.Errorf("you must provide a creator")

	ErrDestroyerNotExist = fmt.Errorf("you must provide a destroyer")

	ErrInvalidMaxIdleSize = fmt.Errorf("invalid maxIdleSize")

	ErrInvalidMaxIdleTime = fmt.Errorf("invalid maxIdleTime")
)

type Pool[T any] interface {
	// this creates or returns a ready-to-use resource from the resource pool
	Acquire(context.Context) (T, error)
	// this releases an active resource back to the resource pool
	Release(context.Context, T)
	// this returns the number of idle resources
	NumIdle() int
}
