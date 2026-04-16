// Package suite_test verifies the behaviour of the suite package.
//
// Covers:
//   - Lifecycle hooks fire in the correct order:
//     SetupSuite → SetupTest → TestX → TearDownTest → TearDownSuite → Shutdown
//   - TearDownSuite always runs before Shutdown (LIFO t.Cleanup ordering)
//   - BaseSuite no-ops allow partial override without compile errors
//   - Multiple Test* methods each get their own SetupTest/TearDownTest wrap
//
// Depends on:
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper/suite
//   - github.com/stretchr/testify/require
//
// Used by:
//   - go test ./suite/...
package suite_test

import (
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/suite"
	"github.com/stretchr/testify/require"
)

// --- Lifecycle order ---------------------------------------------------------

// orderedSuite records the name of each lifecycle event into a shared slice
// so tests can assert the exact firing order.
// The slice is passed in as a pointer so hooks appending inside t.Cleanup
// write to the same slice the assertion reads from.
type orderedSuite struct {
	suite.BaseSuite
	log *[]string // pointer to the slice owned by the test
}

func (s *orderedSuite) SetupSuite()    { *s.log = append(*s.log, "SetupSuite") }
func (s *orderedSuite) SetupTest()     { *s.log = append(*s.log, "SetupTest") }
func (s *orderedSuite) TearDownTest()  { *s.log = append(*s.log, "TearDownTest") }
func (s *orderedSuite) TearDownSuite() { *s.log = append(*s.log, "TearDownSuite") }
func (s *orderedSuite) Shutdown()      { *s.log = append(*s.log, "Shutdown") }
func (s *orderedSuite) TestOne()       { *s.log = append(*s.log, "TestOne") }

func TestSuite_LifecycleOrder_SingleTest(t *testing.T) {
	log := make([]string, 0, 8)
	s := &orderedSuite{log: &log}

	// Register the assertion FIRST so it runs LAST (t.Cleanup is LIFO).
	// By the time this fires, TearDownSuite and Shutdown have already appended
	// their entries — so we see the complete, correct slice.
	t.Cleanup(func() {
		expected := []string{
			"SetupSuite",
			"SetupTest",
			"TestOne",
			"TearDownTest",
			"TearDownSuite",
			"Shutdown",
		}
		require.Equal(t, expected, log,
			"lifecycle hooks did not fire in the expected order")
	})

	suite.Run(t, s)
}

// --- Multiple Test* methods --------------------------------------------------

// multiSuite has two Test* methods to verify each gets its own
// SetupTest / TearDownTest wrap.
type multiSuite struct {
	suite.BaseSuite
	log *[]string
}

func (s *multiSuite) SetupSuite()    { *s.log = append(*s.log, "SetupSuite") }
func (s *multiSuite) SetupTest()     { *s.log = append(*s.log, "SetupTest") }
func (s *multiSuite) TearDownTest()  { *s.log = append(*s.log, "TearDownTest") }
func (s *multiSuite) TearDownSuite() { *s.log = append(*s.log, "TearDownSuite") }
func (s *multiSuite) Shutdown()      { *s.log = append(*s.log, "Shutdown") }
func (s *multiSuite) TestAlpha()     { *s.log = append(*s.log, "TestAlpha") }
func (s *multiSuite) TestBeta()      { *s.log = append(*s.log, "TestBeta") }

func TestSuite_LifecycleOrder_MultipleTests(t *testing.T) {
	log := make([]string, 0, 12)
	s := &multiSuite{log: &log}

	t.Cleanup(func() {
		// SetupSuite fires once. Each Test* gets its own Setup/TearDown wrap.
		// Reflection iterates methods alphabetically, so TestAlpha before TestBeta.
		// TearDownSuite and Shutdown fire once at the very end.
		expected := []string{
			"SetupSuite",
			"SetupTest", "TestAlpha", "TearDownTest",
			"SetupTest", "TestBeta", "TearDownTest",
			"TearDownSuite",
			"Shutdown",
		}
		require.Equal(t, expected, log,
			"lifecycle hooks did not fire in the expected order for multiple tests")
	})

	suite.Run(t, s)
}

// --- Shutdown fires after TearDownSuite --------------------------------------

// shutdownOrderSuite records only TearDownSuite and Shutdown to isolate
// the ordering guarantee between those two specific hooks.
type shutdownOrderSuite struct {
	suite.BaseSuite
	log *[]string
}

func (s *shutdownOrderSuite) TearDownSuite() {
	*s.log = append(*s.log, "TearDownSuite")
}

func (s *shutdownOrderSuite) Shutdown() {
	*s.log = append(*s.log, "Shutdown")
}

// TestSuite_ShutdownAfterTearDownSuite specifically verifies the LIFO
// t.Cleanup guarantee: Shutdown is registered first, so it always runs last.
func TestSuite_ShutdownAfterTearDownSuite(t *testing.T) {
	log := make([]string, 0, 2)
	s := &shutdownOrderSuite{log: &log}

	t.Cleanup(func() {
		require.Equal(t, []string{"TearDownSuite", "Shutdown"}, log,
			"Shutdown must always run after TearDownSuite")
	})

	suite.Run(t, s)
}

// --- BaseSuite no-ops --------------------------------------------------------

// minimalSuite embeds BaseSuite and overrides nothing.
// This verifies that a suite with no overrides compiles and runs without panic.
type minimalSuite struct {
	suite.BaseSuite
}

func (s *minimalSuite) TestNothing() {}

func TestSuite_BaseSuiteNoOpsDoNotPanic(t *testing.T) {
	// If any BaseSuite no-op panics, this test will fail.
	// No assertions needed — a clean run is the signal.
	require.NotPanics(t, func() {
		suite.Run(t, &minimalSuite{})
	})
}
