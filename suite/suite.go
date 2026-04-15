// Package suite provides a struct-based test suite runner with lifecycle hooks.
// Suites embed BaseSuite and override only the hooks they need.
// Run() uses reflection to discover and execute all Test* methods automatically.
//
// Lifecycle order per suite run:
//   SetupSuite()
//   for each Test* method:
//     SetupTest()
//     TestXxx()
//     TearDownTest()
//   TearDownSuite()  ← registered via t.Cleanup (runs after all subtests)
//   Shutdown()       ← registered via t.Cleanup (runs after TearDownSuite)

// Depends on:
//   - testing (stdlib)
//   - reflect (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via suite.Run)
//   - Any consumer project defining struct-based test suites
package suite

import (
	"reflect"
	"strings"
	"testing"
)

// Suite defines the lifecycle interface every test suite must satisfy.
// Embed BaseSuite in your struct to get no-op defaults for all methods,
// then override only the hooks your suite actually needs.
type Suite interface {
	SetupSuite()
	SetupTest()
	TearDownTest()
	TearDownSuite()
	Shutdown()
}

// BaseSuite provides no-op implementations of all Suite lifecycle methods.
// Embedding this in our suite struct allows only to declare the methods we actually want to use.
//
// Example:
//
//	type MySuite struct {
//	    suite.BaseSuite
//	    db *sql.DB
//	}
//
//	func (s *MySuite) SetupSuite() {
//	    s.db = connectDB()
//	}
type BaseSuite struct{}

func (b *BaseSuite) SetupSuite()    {}
func (b *BaseSuite) SetupTest()     {}
func (b *BaseSuite) TearDownTest()  {}
func (b *BaseSuite) TearDownSuite() {}
func (b *BaseSuite) Shutdown()      {}

// Run executes all Test* methods on s as subtests of t, wiring lifecycle hooks automatically around each one.
//
// Shutdown and TearDownSuite are registered via t.Cleanup (not defer) so they
// run even if a test calls t.Fatal or panics. Shutdown is registered first so
// it runs last (t.Cleanup is LIFO), ensuring TearDownSuite always precedes final resource teardown.
//
// Reflection is used only for method discovery: finding methods whose names
// start with "Test" on the concrete type of s. No other reflect usage exists in this package.
func Run(t *testing.T, s Suite) {
	t.Helper()

	// Register Shutdown first - it runs last due to LIFO order
	t.Cleanup(func() {
		s.Shutdown()
	})

	// Register TearDownSuite second - it runs before shutdown.
	t.Cleanup(func() {
		s.TearDownSuite()
	})

	s.SetupSuite()

	// Reflect on the concrete type to discover all Test* methods.
	// We use reflect.TypeOf to iterate method names, then reflect.ValueOf
	// to call them. This is the only use of reflection in the library.
	suiteType := reflect.TypeOf(s)
	suiteValue := reflect.ValueOf(s)

	for i := 0; i < suiteType.NumMethod(); i++ {
		method := suiteType.Method(i)

		if !strings.HasPrefix(method.Name, "Test") {
			continue
		}

		// Capture method name for the closure
		methodName := method.Name
		methodFunc := suiteValue.MethodByName(methodName)

		t.Run(methodName, func(t *testing.T) {
			s.SetupTest()
			methodFunc.Call(nil)
			s.TearDownTest()
		})
	}
}
