package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	mockCTX = context.Background()
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

func TestPoolSuite(t *testing.T) {
	suite.Run(t, new(poolSuite))
}
