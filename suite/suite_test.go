package suite_test

import (
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/suite"
	"github.com/stretchr/testify/assert"
)

type ExampleSuite struct {
	suite.BaseSuite
	setupSuiteDone bool
	setupTestCount int
	teardownTestCount int
	teardownSuiteDone bool
}

func (s *ExampleSuite) SetupSuite() {
	s.setupSuiteDone = true
}

func (s *ExampleSuite) SetupTest() {
	s.setupTestCount++
}

func (s *ExampleSuite) TearDownTest() {
	s.teardownTestCount++
}

func (s *ExampleSuite) TearDownSuite() {
	s.teardownSuiteDone = true
}

func (s *ExampleSuite) TestOne() {
	assert.True(s.T, s.setupSuiteDone)
	assert.Equal(s.T, 1, s.setupTestCount)
}

func (s *ExampleSuite) TestTwo() {
	assert.True(s.T, s.setupSuiteDone)
	assert.Equal(s.T, 2, s.setupTestCount)
}

func TestSuiteRunner(t *testing.T) {
	s := new(ExampleSuite)

	t.Run("RunSuite", func(t *testing.T) {
		suite.Run(t, s)
	})

	if s.teardownTestCount != 2 {
		t.Errorf("expected 2 teardown test calls, got %d", s.teardownTestCount)
	}
	if !s.teardownSuiteDone {
		t.Errorf("teardown suite was not called")
	}
}
