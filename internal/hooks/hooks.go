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
