// Package suite provides a struct-based test suite runner with lifecycle hooks.
// Suites embed BaseSuite and override only the hooks they need.
// Run() uses reflection to discover and execute all Test* methods automatically.
//
// Depends on:
//   - testing (stdlib)
//   - reflect (stdlib)
//
// Used by:
//   - testifywrapper.Instance (via suite.Run)
//   - Any consumer project defining struct-based test suites
package suite

import (
	"reflect"
	"strings"
	"testing"
)

// BaseSuite is the base struct for all test suites.
type BaseSuite struct {
	T *testing.T
}

// SetT sets the *testing.T for the suite.
func (s *BaseSuite) SetT(t *testing.T) {
	s.T = t
}

// SetupAllSuite runs once before the entire suite.
type SetupAllSuite interface {
	SetupSuite()
}

// SetupTestSuite runs before each test in the suite.
type SetupTestSuite interface {
	SetupTest()
}

// TearDownTestSuite runs after each test in the suite.
type TearDownTestSuite interface {
	TearDownTest()
}

// TearDownAllSuite runs once after the entire suite.
type TearDownAllSuite interface {
	TearDownSuite()
}

// Run executes all Test* methods in the given suite.
func Run(t *testing.T, suite interface{}) {
	t.Helper()

	suiteValue := reflect.ValueOf(suite)
	suiteType := suiteValue.Type()

	// Set T if it's a BaseSuite or has SetT
	if setter, ok := suite.(interface{ SetT(*testing.T) }); ok {
		setter.SetT(t)
	}

	// SetupSuite
	if setupSuite, ok := suite.(SetupAllSuite); ok {
		setupSuite.SetupSuite()
	}

	// TearDownSuite via Cleanup
	if tearDownSuite, ok := suite.(TearDownAllSuite); ok {
		t.Cleanup(func() {
			tearDownSuite.TearDownSuite()
		})
	}

	for i := 0; i < suiteType.NumMethod(); i++ {
		method := suiteType.Method(i)
		if strings.HasPrefix(method.Name, "Test") {
			// Validate method signature: func (s *MySuite) TestSomething()
			// It should have 1 input (the receiver) and 0 outputs.
			if method.Type.NumIn() != 1 || method.Type.NumOut() != 0 {
				continue
			}

			t.Run(method.Name, func(t *testing.T) {
				// Re-set T for the subtest
				if setter, ok := suite.(interface{ SetT(*testing.T) }); ok {
					setter.SetT(t)
				}

				// SetupTest
				if setupTest, ok := suite.(SetupTestSuite); ok {
					setupTest.SetupTest()
				}

				// TearDownTest via Cleanup
				if tearDownTest, ok := suite.(TearDownTestSuite); ok {
					t.Cleanup(func() {
						tearDownTest.TearDownTest()
					})
				}

				// Call the test method
				method.Func.Call([]reflect.Value{suiteValue})
			})
		}
	}
}
