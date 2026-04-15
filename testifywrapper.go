// Package testifywrapper provides a single, consistent testing harness for Go projects.
// It wraps testify with lifecycle management, file-driven test support, and suite running —
// all accessible through one New(t) call.
//
// Depends on:
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper/assert
//   - github.com/FarhanAsfar/testify-wrapper/require
//   - github.com/FarhanAsfar/testify-wrapper/filehandler
//   - github.com/FarhanAsfar/testify-wrapper/internal/hooks
//
// Used by:
//   - Any Go project as its primary test harness
package testifywrapper

import (
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/assert"
	"github.com/FarhanAsfar/testify-wrapper/filehandler"
	"github.com/FarhanAsfar/testify-wrapper/internal/hooks"
	"github.com/FarhanAsfar/testify-wrapper/require"
)

// Instance provides access to all testing tools in the harness.
type Instance struct {
	t *testing.T
}

// New creates a new Instance bound to the given *testing.T.
func New(t *testing.T) *Instance {
	t.Helper()
	return &Instance{t: t}
}

// Assert returns an assert.Assertions instance.
func (i *Instance) Assert() *assert.Assertions {
	i.t.Helper()
	return assert.New(i.t)
}

// Require returns a require.Assertions instance.
func (i *Instance) Require() *require.Assertions {
	i.t.Helper()
	return require.New(i.t)
}

// FileHandler returns a filehandler.Handler instance.
func (i *Instance) FileHandler() *filehandler.Handler {
	i.t.Helper()
	return filehandler.New(i.t)
}

// Hooks returns a hooks.Registry instance.
func (i *Instance) Hooks() *hooks.Registry {
	i.t.Helper()
	return hooks.New(i.t)
}
