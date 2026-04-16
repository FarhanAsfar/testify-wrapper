// Package testifyWrapper provides a single, consistent testing harness for Go projects.
// It wraps testify with lifecycle management, file-driven test support, and suite running —
// all accessible through one New(t) call.
//
// Typical usage:
//
//	func TestSomething(t *testing.T) {
//	    kit := testifyWrapper.New(t)
//	    kit.Assert().Equal(42, result)
//	    kit.Require().NoError(err)
//	}
//
// File-driven usage:
//
//	func TestWithFixtures(t *testing.T) {
//	    kit := testifyWrapper.New(t)
//	    cases, err := kit.LoadJSON("testdata/cases.json")
//	    kit.Require().NoError(err)
//	    kit.RunCases(t, cases, func(t *testing.T, tc testifyWrapper.TestCase) {
//	        // unmarshal tc.Input and tc.Expected, run logic, assert results
//	    })
//	}
//
// Depends on:
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper/assert
//   - github.com/FarhanAsfar/testify-wrapper/filehandler
//   - github.com/FarhanAsfar/testify-wrapper/internal/hooks
//   - github.com/FarhanAsfar/testify-wrapper/require
//
// Used by:
//   - Any Go project as its primary test harness
package testifyWrapper

import (
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/assert"
	"github.com/FarhanAsfar/testify-wrapper/filehandler"
	"github.com/FarhanAsfar/testify-wrapper/internal/hooks"
	"github.com/FarhanAsfar/testify-wrapper/require"
)

// TestCase represents a single test case loaded from a JSON or YAML fixture file.
//
// It is re-exported from the filehandler package for convenience, so consumers
// can use testifyWrapper.TestCase without importing filehandler directly.
//
// Schema:
//
//	{
//	    "name":     "descriptive name used as the subtest label",
//	    "input":    { ...your input structure... },
//	    "expected": { ...your expected output structure... }
//	}
//
// Input and Expected are raw JSON — unmarshal them into your own concrete types
// inside the RunCases callback.
type TestCase = filehandler.TestCase

// Instance is the central object for a single test function.
// It owns the assert wrapper, require wrapper, file handler, and cleanup registry
// for its *testing.T. One Instance per test — never share across tests.
//
// Obtain an Instance by calling New(t) at the top of your test function.
type Instance struct {
	// T is the *testing.T for the current test. Exposed for cases where
	// the consumer needs to pass it to a helper that does not accept Instance.
	T *testing.T

	asserter    *assert.Assertions
	requirer    *require.Assertions
	fileHandler *filehandler.FileHandler
	hookReg     *hooks.Registry
}

// New creates an Instance for the given *testing.T.
// Call this once at the top of each test function.
// The returned Instance must not be stored in any variable that outlives the test.
func New(t *testing.T) *Instance {
	t.Helper()
	return &Instance{
		T:           t,
		asserter:    assert.New(t),
		requirer:    require.New(t),
		fileHandler: filehandler.New(),
		hookReg:     hooks.New(t),
	}
}

// Assert returns the non-fatal assertion instance for this test.
// Failures are recorded but do not stop test execution.
// Use Require() when a failure should halt the test immediately.
func (i *Instance) Assert() *assert.Assertions {
	return i.asserter
}

// Require returns the fatal assertion instance for this test.
// Failures immediately stop test execution via t.FailNow().
// Use Assert() when you want the test to continue collecting failures.
func (i *Instance) Require() *require.Assertions {
	return i.requirer
}

// FileHandler returns the underlying FileHandler for this test.
// Use this when you need direct access to the filehandler API.
// For the common case, prefer the LoadJSON, LoadYAML, and RunCases
// convenience methods on Instance directly.
func (i *Instance) FileHandler() *filehandler.FileHandler {
	return i.fileHandler
}

// LoadJSON reads the file at path and parses it as a JSON array of TestCase.
// Returns a typed error (filehandler.ErrFileNotFound or filehandler.ErrInvalidFormat)
// so callers can use errors.Is() for specific handling.
//
// Pair with RunCases to execute each case as a named subtest:
//
//	cases, err := kit.LoadJSON("testdata/cases.json")
//	kit.Require().NoError(err)
//	kit.RunCases(t, cases, func(t *testing.T, tc testifyWrapper.TestCase) { ... })
func (i *Instance) LoadJSON(path string) ([]TestCase, error) {
	return i.fileHandler.LoadJSON(path)
}

// LoadYAML reads the file at path and parses it as a YAML array of TestCase.
// Input and Expected fields are normalised to json.RawMessage so the RunCases
// callback is identical regardless of whether cases came from JSON or YAML.
// Returns a typed error (filehandler.ErrFileNotFound or filehandler.ErrInvalidFormat).
func (i *Instance) LoadYAML(path string) ([]TestCase, error) {
	return i.fileHandler.LoadYAML(path)
}

// RunCases executes each TestCase as a named subtest via t.Run.
// The fn callback receives the subtest's *testing.T and the current TestCase.
// testifyWrapper owns the loop and subtest wiring — fn owns the assertion logic.
//
// Cases with an empty Name are run with a generated fallback name ("case_0", "case_1", ...).
// Always pair the subtest's t with its own New(t) call inside fn:
//
//	kit.RunCases(t, cases, func(t *testing.T, tc testifyWrapper.TestCase) {
//	    subKit := testifyWrapper.New(t)
//	    subKit.Assert().Equal(expected, actual)
//	})
func (i *Instance) RunCases(t *testing.T, cases []TestCase, fn func(t *testing.T, tc TestCase)) {
	i.fileHandler.RunCases(t, cases, fn)
}

// RegisterCleanup enqueues fn to run when the test completes.
// Functions are executed in LIFO order. fn must be non-nil.
// Prefer this over defer for all teardown inside test code —
// t.Cleanup runs even on t.Fatal and panics.
func (i *Instance) RegisterCleanup(fn func()) {
	i.hookReg.Register(fn)
}
