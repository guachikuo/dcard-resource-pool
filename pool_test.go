package pool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	mockCTX = context.Background()

	defaultMaxIdleSize = 10
)

type poolSuite struct {
	suite.Suite
}

func (p *poolSuite) SetupSuite() {
}

func (p *poolSuite) TearDownSuite() {
}

func (p *poolSuite) SetupTest() {
}

func (p *poolSuite) TearDownTest() {
}

func (p *poolSuite) TestNew() {
	tests := []struct {
		desc string

		creator     func(context.Context) (any, error)
		destroyer   func(context.Context, any)
		maxIdleSize int
		maxIdleTime time.Duration

		err error
	}{
		{
			desc:    "not providing creator",
			creator: nil,
			err:     ErrCreatorNotExist,
		},
		{
			desc:      "not providing destroyer",
			creator:   func(context.Context) (any, error) { return *new(any), nil },
			destroyer: nil,
			err:       ErrDestroyerNotExist,
		},
		{
			desc:        "maxIdleSize is 0 which is invalid",
			creator:     func(context.Context) (any, error) { return *new(any), nil },
			destroyer:   func(context.Context, any) {},
			maxIdleSize: 0,
			err:         ErrInvalidMaxIdleSize,
		},
		{
			desc:        "maxIdleSize is -1 which is invalid",
			creator:     func(context.Context) (any, error) { return *new(any), nil },
			destroyer:   func(context.Context, any) {},
			maxIdleSize: -1,
			err:         ErrInvalidMaxIdleSize,
		},
		{
			desc:        "maxIdleTime is 0 which is invalid",
			creator:     func(context.Context) (any, error) { return *new(any), nil },
			destroyer:   func(context.Context, any) {},
			maxIdleSize: 10,
			maxIdleTime: time.Duration(0),
			err:         ErrInvalidMaxIdleTime,
		},
		{
			desc:        "maxIdleTime is -1 which is invalid",
			creator:     func(context.Context) (any, error) { return *new(any), nil },
			destroyer:   func(context.Context, any) {},
			maxIdleSize: 10,
			maxIdleTime: time.Duration(-1),
			err:         ErrInvalidMaxIdleTime,
		},
		{
			desc: "error ocurred when creating resource",
			creator: func(context.Context) (any, error) {
				return *new(any), fmt.Errorf("error ocurred when creating resource")
			},
			destroyer:   func(context.Context, any) {},
			maxIdleSize: 10,
			maxIdleTime: time.Duration(1),
			err:         fmt.Errorf("error ocurred when creating resource"),
		},
		{
			desc:        "success",
			creator:     func(context.Context) (any, error) { return *new(any), nil },
			destroyer:   func(context.Context, any) {},
			maxIdleSize: 10,
			maxIdleTime: time.Duration(1),
			err:         nil,
		},
	}

	for _, test := range tests {
		_, err := New[any](
			test.creator,
			test.destroyer,
			test.maxIdleSize,
			test.maxIdleTime,
		)

		p.Require().Equal(test.err, err)
	}
}

func (p *poolSuite) NewAny() Pool[any] {
	pool, err := New[any](
		func(context.Context) (any, error) { return *new(any), nil },
		func(context.Context, any) {},
		defaultMaxIdleSize,
		time.Duration(1*time.Second),
	)
	p.Require().NoError(err)

	return pool
}

func (p *poolSuite) TestAcquireReturnReadyToUseResource() {
	pool := p.NewAny()

	t, err := pool.Acquire(mockCTX)
	p.Require().NoError(err)
	p.Require().Empty(t)

	num := pool.NumIdle()
	p.Require().Equal(defaultMaxIdleSize-1, num)
}

func (p *poolSuite) TestAcquireReturnNewResource() {
	pool := p.NewAny()

	// sleep 1 second,
	// waiting for resource in pool is alive for more than maxIdleTime
	time.Sleep(time.Second)

	t, err := pool.Acquire(mockCTX)
	p.Require().NoError(err)
	p.Require().Empty(t)

	num := pool.NumIdle()
	p.Require().Equal(0, num)
}

func (p *poolSuite) TestRelease() {
	pool := p.NewAny()

	for i := 0; i < 3; i++ {
		t, err := pool.Acquire(mockCTX)
		p.Require().NoError(err)
		p.Require().Empty(t)
	}

	pool.Release(mockCTX, *new(any))

	// 10 - 3 + 1 resource
	num := pool.NumIdle()
	p.Require().Equal(8, num)
}

func (p *poolSuite) TestNumIdle() {
	pool := p.NewAny()

	fn := func(wg *sync.WaitGroup) {
		defer wg.Done()
		num := pool.NumIdle()

		p.Require().Equal(defaultMaxIdleSize, num)
	}

	goCnt := 10
	wg := new(sync.WaitGroup)
	for i := 0; i < goCnt; i++ {
		wg.Add(1)
		go fn(wg)
	}
	wg.Wait()
}

func (p *poolSuite) TestByRunningSmallIntegration() {
	pool := p.NewAny()

	fn := func(i int, wg *sync.WaitGroup) {
		defer wg.Done()

		t, err := pool.Acquire(mockCTX)
		p.Require().NoError(err)
		p.Require().Empty(t)

		if i%3 == 0 {
			time.Sleep(100 * time.Millisecond)
			pool.Release(mockCTX, t)
		}
	}

	wg := new(sync.WaitGroup)
	goCnt := 18
	for i := 0; i < goCnt; i++ {
		wg.Add(1)
		go fn(i, wg)
	}
	wg.Wait()

	// acquire but not release : 18*(2/3)
	// acquire and release : 18*(1/3)
	num := pool.NumIdle()
	p.Require().Equal(6, num)

	wg = new(sync.WaitGroup)
	goCnt = 100
	for i := 0; i < goCnt; i++ {
		wg.Add(1)
		go fn(i, wg)
	}
	wg.Wait()

	// acquire but not release : 100*(2/3)
	// acquire and release : 100*(1/3), but only 10 resource would be in the pool
	num = pool.NumIdle()
	p.Require().Equal(10, num)
}

func TestPoolSuite(t *testing.T) {
	suite.Run(t, new(poolSuite))
}
