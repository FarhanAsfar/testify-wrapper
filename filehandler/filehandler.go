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
//   - testifywrapper.Instance (via kit.FileHandler())
//   - Any consumer project running data-driven tests
package filehandler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestCase represents a single test case loaded from a file.
type TestCase struct {
	Name     string                 `json:"name" yaml:"name"`
	Input    map[string]interface{} `json:"input" yaml:"input"`
	Expected map[string]interface{} `json:"expected" yaml:"expected"`
}

// Handler manages loading and running test cases.
type Handler struct {
	t *testing.T
}

// New creates a new Handler instance.
func New(t *testing.T) *Handler {
	t.Helper()
	return &Handler{t: t}
}

// RunCases loads test cases from the specified file and runs them as subtests.
func (h *Handler) RunCases(filename string, testFn func(t *testing.T, tc TestCase)) error {
	h.t.Helper()

	cases, err := h.LoadCases(filename)
	if err != nil {
		return err
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		h.t.Run(tc.Name, func(t *testing.T) {
			testFn(t, tc)
		})
	}

	return nil
}

// LoadCases reads test cases from a JSON or YAML file.
func (h *Handler) LoadCases(filename string) ([]TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cases []TestCase
	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cases); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cases); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return cases, nil
}
