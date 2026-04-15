// Package require provides a thin, opinionated wrapper around testify/require.
// Assertion failures immediately stop test execution via t.FailNow().
// Use this when subsequent test steps would be meaningless or dangerous after a failure.

// By embedding *require.Assertions, all standard testify assertion methods
// (Equal, Nil, NoError, etc.) are available directly on Assertions.
// We can add any custom require-level assertions as methods on this type.

// Depends on:
//   - github.com/stretchr/testify/require
//   - testing (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via kit.Require())
//   - Any consumer project that imports testifyWrapper

package require

import (
	"testing"

	testifyrequire "github.com/stretchr/testify/require"
)

// Assertions wraps testify's require.Assertions.
// All standard testify assertion methods are promoted via embedding.
// Failures are fatal: the test stops immediately at the point of failure.
type Assertions struct {
	*testifyrequire.Assertions
}

// New creates an Assertions instance bound to the given *testing.T.
// The returned instance must not be shared across goroutines or tests.
func New(t *testing.T) *Assertions {
	t.Helper()
	return &Assertions{
		Assertions: testifyrequire.New(t),
	}
}
