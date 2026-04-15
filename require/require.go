// Package require provides a thin, opinionated wrapper around testify/require.
// Assertion failures immediately stop test execution via t.FailNow().
// Use this when subsequent test steps would be meaningless or dangerous after a failure.
//
// Depends on:
//   - github.com/stretchr/testify/require
//   - testing (stdlib)
//
// Used by:
//   - testifywrapper.Instance (via kit.Require())
//   - Any consumer project that imports testifywrapper
package require

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Assertions provides assertion methods that stop test execution on failure.
type Assertions struct {
	t *testing.T
}

// New creates a new Assertions instance bound to the given *testing.T.
func New(t *testing.T) *Assertions {
	t.Helper()
	return &Assertions{t: t}
}

// Equal asserts that two objects are equal.
func (r *Assertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	require.Equal(r.t, expected, actual, msgAndArgs...)
}

// NoError asserts that a function returned no error.
func (r *Assertions) NoError(err error, msgAndArgs ...interface{}) {
	r.t.Helper()
	require.NoError(r.t, err, msgAndArgs...)
}

// True asserts that the specified value is true.
func (r *Assertions) True(value bool, msgAndArgs ...interface{}) {
	r.t.Helper()
	require.True(r.t, value, msgAndArgs...)
}

// NotNil asserts that the specified object is not nil.
func (r *Assertions) NotNil(object interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	require.NotNil(r.t, object, msgAndArgs...)
}

// Nil asserts that the specified object is nil.
func (r *Assertions) Nil(object interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	require.Nil(r.t, object, msgAndArgs...)
}
