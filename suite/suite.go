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

	// T returns the *testing.T for the currently running Test* method.
	// It is set by the runner before each SetupTest() call and is valid for the full duration of SetupTest -> TestXxx -> TearDownTest.
	T() *testing.T

	// setT is unexported - only the runner calls it.
	// It is defined on BaseSuite and must not be overridden by consumers.
	setT(t *testing.T)
}

// BaseSuite provides no-op implementations of all Suite lifecycle methods and manages the current *testing.T for the running test method.
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
//	func (s *MyServiceSuite) Shutdown()   { s.db.Close() }
//
//	func (s *MyServiceSuite) TestCreate() {
//	    kit := testifyWrapper.New(s.T())
//	    kit.Assert().NoError(s.db.Ping())
//	}
type BaseSuite struct {
	// currentT holds the *testing.T for the currently running Test* method.
	// It is set by the runner via setT() before each test method runs.
	// Unexported so consumers cannot accidentally overwrite it.
	currentT *testing.T
}

// T returns the *testing.T for the currently running Test* method.
// T() returns nil if called outside of a running test method (e.g. inside
// SetupSuite or Shutdown) — those hooks run outside the subtest scope.
func (b *BaseSuite) T() *testing.T {
	return b.currentT
}

// setT is called by the runner to bind the current subtest's *testing.T
// before each SetupTest → TestXxx → TearDownTest cycle.
// Consumers must not call or override this method.
func (b *BaseSuite) setT(t *testing.T) {
	b.currentT = t
}

// SetupSuite is a no-op. Override in suite to run once before all tests.
func (b *BaseSuite) SetupSuite() {}

// SetupTest is a no-op. Override in suite to run before each Test* method.
func (b *BaseSuite) SetupTest() {}

// TearDownTest is a no-op. Override in suite to run after each Test* method.
func (b *BaseSuite) TearDownTest() {}

// TearDownSuite is a no-op. Override in suite to run once after all tests.
func (b *BaseSuite) TearDownSuite() {}

// Shutdown is a no-op. Override in suite to release long-lived resources.
// Shutdown always runs after TearDownSuite — it is the guaranteed-final hook.
func (b *BaseSuite) Shutdown() {}

// Run executes all Test* methods on s as subtests of t, wiring lifecycle hooks automatically around each one.
//
// Before each Test* method, the subtest's *testing.T is stored on the suite
// via setT() so suite methods can access it through s.T().

// Shutdown and TearDownSuite are registered via t.Cleanup (not defer) so they
// run even if a test calls t.Fatal or panics. Shutdown is registered first so
// it runs last (t.Cleanup is LIFO), ensuring TearDownSuite always precedes final resource teardown.
//
// Reflection is used only for method discovery: finding methods whose names
// start with "Test" on the concrete type of s. Only methods with no parameters
// (beyond the receiver) and no return values are executed — others are logged
// and skipped.
//
//	No other reflect usage exists in this package.
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

		// Validate signature strictly: no parameters beyond the receiver,
		// no return values. method.Type includes the receiver as the first
		// parameter, so a valid Test* method has NumIn() == 1, NumOut() == 0.
		// Log and skip anything that does not match — never panic.
		if method.Type.NumIn() != 1 || method.Type.NumOut() != 0 {
			t.Logf("suite: skipping %s — must have no parameters and no return values", method.Name)
			continue
		}

		// Capture method name for the closure
		methodName := method.Name
		methodFunc := suiteValue.MethodByName(methodName)

		t.Run(methodName, func(t *testing.T) {
			// Bind the subtest's *testing.T to the suite before SetupTest
			// so it is available for the full SetupTest → TestXxx → TearDownTest cycle.
			s.setT(t)
			s.SetupTest()
			methodFunc.Call(nil)
			s.TearDownTest()
		})
	}
}
