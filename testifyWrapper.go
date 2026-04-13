// Package testifyWrapper provides a single, consistent testing harness for Go projects.
// It wraps testify with lifecycle management, file-driven test support, and suite running —
// all accessible through one New(t) call.
//
// Depends on:
//   - testing (stdlib)
//   - github.com/w3engineers/testifyWrapper/assert
//   - github.com/w3engineers/testifyWrapper/require
//   - github.com/w3engineers/testifyWrapper/filehandler
//
// Used by:
//   - Any Go project as its primary test harness
package testifyWrapper
