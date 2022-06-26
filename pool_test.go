package pool

import (
	"context"
	"testing"

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

func TestPoolSuite(t *testing.T) {
}
