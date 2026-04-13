// Package suite provides a struct-based test suite runner with lifecycle hooks.
// Suites embed BaseSuite and override only the hooks they need.
// Run() uses reflection to discover and execute all Test* methods automatically.
//
// Depends on:
//   - testing (stdlib)
//   - reflect (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via suite.Run)
//   - Any consumer project defining struct-based test suites
package suite
