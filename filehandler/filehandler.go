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

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert/yaml"
)

// Sentinel errors for typed error checking via errors.Is()
var (
	// ErrFileNotFound is returned when the fixture file path does not exist.
	ErrFileNotFound = errors.New("Fixture file not found")

	// ErrInvalidFormat is returned when the file content can't be parsed into the expected []TestCase structure
	ErrInvalidFormat = errors.New("Fixture file has invalid format")
)

// TestCase is the schema every entry in a fixture file must conform to.
// Name is used as the subtest name in t.Run.
// Input and Expected are raw JSON — consumers unmarshal into their own concrete types.
type TestCase struct {
	Name     string          `json:"name"		yaml:"name"`
	Input    json.RawMessage `json:"input"		yaml:"input"`
	Expected json.RawMessage `json:"expected"		yaml:"expected"`
}

// FileHandler loads and runs fixture-based test cases.
// Each FileHandler is stateless — the same instance can load multiple files.
// It is safe to share across subtests since it holds no mutable state.
type FileHandler struct{}

// New creates a new FileHanlder
// FileHandler is stateless, but New() is provided for consistency with the rest of the testifyWrapper API and to allow future extension without breaking the call sites
func New() *FileHandler {
	return &FileHandler{}
}

// LoadJSON reads the file at path and parses it as a JSON array of TestCase.
// Returns ErrFileNotFound if the path doesn't exist.
// Returns ErrInvalidFormat if the content is not a valid []TestCase.
func (fh *FileHandler) LoadJSON(path string) ([]TestCase, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, fmt.Errorf("%w: %s %v", ErrInvalidFormat, path, err)
	}
	return cases, nil
}

// LoadYAML reads the file at path and parses it as a YAML array of TestCase.
// Input and Expected fields are normalised to json.RawMessage so consumers
// always work with the same type regardless of the source file format.
// Returns ErrFileNotFound if the path does not exist.
// Returns ErrInvalidFormat if the content is not a valid []TestCase structure.
func (fh *FileHandler) LoadYAML(path string) ([]TestCase, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	// yamlTestCase mirrors TestCase but uses interface{} for Input and Expected
	// because yaml.v3 has no RawMessage equivalent. After unmarshalling we
	// re-marshal each field to JSON bytes so TestCase always carries json.RawMessage.
	type yamlTestCase struct {
		Name     string `yaml:"name"`
		Input    any    `yaml:"input"`
		Expected any    `yaml:"expected"`
	}

	var raw []yamlTestCase
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInvalidFormat, path, err)
	}

	cases := make([]TestCase, 0, len(raw))
	for i, r := range raw {
		inputJSON, err := json.Marshal(r.Input)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: case %d input: %v", ErrInvalidFormat, path, i, err)
		}

		expectedJSON, err := json.Marshal(r.Expected)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: case %d expected: %v", ErrInvalidFormat, path, i, err)
		}

		cases = append(cases, TestCase{
			Name:     r.Name,
			Input:    json.RawMessage(inputJSON),
			Expected: json.RawMessage(expectedJSON),
		})
	}

	return cases, nil
}

// RunCases executes each TestCase as a named subtest via t.Run.
// The provided fn callback receives the subtest's *testing.T and the current TestCase.
// testifyWrapper owns the loop and subtest wiring — fn owns the assertion logic.
// Cases with an empty Name are run with a generated fallback name ("case_0", "case_1", ...).
func (fh *FileHandler) RunCases(t *testing.T, cases []TestCase, fn func(t *testing.T, tc TestCase)) {
	t.Helper()

	for i, tc := range cases {
		tc := tc // capture loop vairable for safe use in subtest closure

		name := tc.Name
		if name == "" {
			name = fmt.Sprintf("case_%d", i)
		}

		t.Run(name, func(t *testing.T) {
			fn(t, tc)
		})
	}
}

// readFile is a shared helper for reading fixture files from disk.
// It wraps os errors into the appropriate sentinel types.
func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		// Other OS errors (permission denied, etc.) are wrapped with context
		// but not mapped to a sentinel — they are infrastructure issues, not format issues.
		return nil, fmt.Errorf("could not read fixture file %s: %w", path, err)
	}
	return data, nil
}
