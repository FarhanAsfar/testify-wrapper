// Package suite_test contains tests for the suite package.
// It verifies lifecycle hook ordering and that Shutdown fires correctly via t.Cleanup.
//
// Depends on:
//   - testing (stdlib)
//   - github.com/w3engineers/testifyWrapper/suite
//
// Used by:
//   - go test ./suite/...
package suite_test
