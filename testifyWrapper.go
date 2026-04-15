// Package testifyWrapper provides a single, consistent testing harness for Go projects.
// It wraps testify with lifecycle management, file-driven test support, and suite running —
// all accessible through one New(t) call.

//Typical usage:
//	func TestSomething(t *testing.T) {
//	    kit := testifyWrapper.New(t)
//	    kit.Assert().Equal(42, result)
//	    kit.Require().NoError(err)
//	}

// Depends on:
//   - testing (stdlib)
//   - github.com/w3engineers/testifyWrapper/assert
//   - github.com/w3engineers/testifyWrapper/require
//   - github.com/w3engineers/testifyWrapper/filehandler
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

// Instance is the central object for a single test function.
// It owns the assert wrapper, require wrapper, file handler, and cleanup registry
// for its *testing.T. One Instance per test — never share across tests.
//
// Obtain an Instance by calling New(t) at the top of your test function.
type Instance struct {
	// T is the *testing.T for the current test. Exposed for cases where the consumer needs to pass it to a helper that doesn't accept Instance.
	T *testing.T

	asserter    *assert.Assertions
	requirer    *require.Assertions
	fileHandler *filehandler.FileHandler
	hookReg     *hooks.Registry
}

// New creates an Instance for the given *testing.T
// Call this once at the top of each test function
// The returned Instance must not be stored in any variable that outlives the test
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

// FileHandler returns the file handler for this test.
// Use it to load JSON or YAML fixture files and run data-driven subtests.
func (i *Instance) FileHandler() *filehandler.FileHandler {
	return i.fileHandler
}

// RegisterCleanup enqueues fn to run when the test completes.
// Functions are executed in LIFO order. fn must be non-nil.
// Prefer this over defer for all teardown inside test code —
// t.Cleanup runs even on t.Fatal and panics.
func (i *Instance) RegisterCleanup(fn func()) {
	i.hookReg.Register(fn)
}
