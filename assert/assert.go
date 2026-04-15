// Package assert provides a thin, opinionated wrapper around testify/assert.
// Assertion failures are recorded but do not stop test execution — the test continues and collects all failures.
//
// By embedding *assert.Assertions, all standard testify assertion methods
// (Equal, Nil, NoError, etc.) are available directly on Assertions.
// we can add any custom assertions as methods on this type.

// Depends on:
//   - github.com/stretchr/testify/assert
//   - testing (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via kit.Assert())
//   - Any consumer project that imports testifyWrapper
package assert

import (
	"testing"

	testifyassert "github.com/stretchr/testify/assert"
)

// Assertions wraps testify's assert.Assertions.
// All standard testify assertion methods are promoted via embedding.
// Failures are non-fatal: the test is marked as failed but continues running.

type Assertions struct {
	*testifyassert.Assertions
}

// New creates an Assertions instance bound to the given *testing.T.
// The returned instance must not be shared across goroutines or tests.
func New(t *testing.T) *Assertions {
	t.Helper()
	return &Assertions{
		Assertions: testifyassert.New(t),
	}
}
