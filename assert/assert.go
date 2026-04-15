// Package assert provides a thin, opinionated wrapper around testify/assert.
// Assertion failures are recorded but do not stop test execution — the test continues and collects all failures.
//
// Depends on:
//   - github.com/stretchr/testify/assert
//   - testing (stdlib)
//
// Used by:
//   - testifywrapper.Instance (via kit.Assert())
//   - Any consumer project that imports testifywrapper
package assert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Assertions provides assertion methods that record failures but allow test execution to continue.
type Assertions struct {
	t *testing.T
}

// New creates a new Assertions instance bound to the given *testing.T.
func New(t *testing.T) *Assertions {
	t.Helper()
	return &Assertions{t: t}
}

// Equal asserts that two objects are equal.
func (a *Assertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return assert.Equal(a.t, expected, actual, msgAndArgs...)
}

// NoError asserts that a function returned no error.
func (a *Assertions) NoError(err error, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return assert.NoError(a.t, err, msgAndArgs...)
}

// True asserts that the specified value is true.
func (a *Assertions) True(value bool, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return assert.True(a.t, value, msgAndArgs...)
}

// NotNil asserts that the specified object is not nil.
func (a *Assertions) NotNil(object interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return assert.NotNil(a.t, object, msgAndArgs...)
}

// Nil asserts that the specified object is nil.
func (a *Assertions) Nil(object interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return assert.Nil(a.t, object, msgAndArgs...)
}
