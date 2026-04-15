// Package hooks provides an internal cleanup hook registry for testifyWrapper.
// It allows registering cleanup functions that are flushed in LIFO order,
// ensuring resources are released in the reverse order they were acquired.
//
// Depends on:
//   - testing (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (shutdown/cleanup wiring)
package hooks

import "testing"

// Registry holds a reference to the test's *testing.T and provides
// a single method to register cleanup functions against it.
// Each Registry is scoped to one *testing.T and must not be shared across tests.
type Registry struct {
	t *testing.T
}

// New creates a Registry bound to the given *testing.T
func New(t *testing.T) *Registry {
	t.Helper()
	return &Registry{t: t}
}

// Register enqueues fn to run when the test completes.
// Multiple registrations execute in LIFO order (last registered, first executed).
// fn must be non-nil.
func (r *Registry) Register(fn func()) {
	r.t.Helper()
	if fn == nil {
		return
	}
	r.t.Cleanup(fn)
}
