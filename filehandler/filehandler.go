// Package filehandler provides data-driven test support by loading test cases
// from JSON or YAML fixture files. Each file must conform to the TestCase schema.
// Errors are always returned as typed sentinel errors — never swallowed silently.
//
// Depends on:
//   - encoding/json (stdlib)
//   - os (stdlib)
//   - gopkg.in/yaml.v3
//
// Used by:
//   - testifyWrapper.Instance (via kit.FileHandler())
//   - Any consumer project running data-driven tests
package filehandler
