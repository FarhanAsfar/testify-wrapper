// Package require provides a thin, opinionated wrapper around testify/require.
// Assertion failures immediately stop test execution via t.FailNow().
// Use this when subsequent test steps would be meaningless or dangerous after a failure.
//
// Depends on:
//   - github.com/stretchr/testify/require
//   - testing (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via kit.Require())
//   - Any consumer project that imports testifyWrapper
package require
