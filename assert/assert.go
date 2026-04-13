// Package assert provides a thin, opinionated wrapper around testify/assert.
// Assertion failures are recorded but do not stop test execution — the test continues and collects all failures.
//
// Depends on:
//   - github.com/stretchr/testify/assert
//   - testing (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via kit.Assert())
//   - Any consumer project that imports testifyWrapper
package assert
