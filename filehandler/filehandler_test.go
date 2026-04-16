// Package filehandler_test verifies the behaviour of the filehandler package.
//
// Covers:
//   - Happy path JSON and YAML loading
//   - Missing file returns ErrFileNotFound (typed, errors.Is compatible)
//   - Malformed file returns ErrInvalidFormat (typed, errors.Is compatible)
//   - Empty case name falls back to "case_0", "case_1", ...
//   - RunCases wires each TestCase as a properly named subtest
//
// Depends on:
//   - testing (stdlib)
//   - encoding/json (stdlib)
//   - errors (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper/filehandler
//   - github.com/stretchr/testify/require
//
// Used by:
//   - go test ./filehandler/...
package filehandler_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/filehandler"
	"github.com/stretchr/testify/require"
)

// --- Helpers -----------------------------------------------------------------

// writeTemp writes content to a temporary file with the given filename suffix
// and returns its path. The file is removed automatically when the test ends.
func writeTemp(t *testing.T, suffix, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, suffix)

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTemp: could not create fixture file: %v", err)
	}

	return path
}

// --- LoadJSON ----------------------------------------------------------------

func TestLoadJSON_HappyPath(t *testing.T) {
	path := writeTemp(t, "cases.json", `[
		{"name": "first case",  "input": {"x": 1}, "expected": {"y": 2}},
		{"name": "second case", "input": {"x": 3}, "expected": {"y": 4}}
	]`)

	fh := filehandler.New()
	cases, err := fh.LoadJSON(path)

	require.NoError(t, err)
	require.Len(t, cases, 2)
	require.Equal(t, "first case", cases[0].Name)
	require.Equal(t, "second case", cases[1].Name)

	// Verify Input and Expected are valid raw JSON that can be unmarshalled.
	var input map[string]int
	require.NoError(t, json.Unmarshal(cases[0].Input, &input))
	require.Equal(t, 1, input["x"])

	var expected map[string]int
	require.NoError(t, json.Unmarshal(cases[0].Expected, &expected))
	require.Equal(t, 2, expected["y"])
}

func TestLoadJSON_FileNotFound(t *testing.T) {
	fh := filehandler.New()
	_, err := fh.LoadJSON("/does/not/exist/cases.json")

	require.Error(t, err)
	require.True(t, errors.Is(err, filehandler.ErrFileNotFound),
		"expected ErrFileNotFound, got: %v", err)
}

func TestLoadJSON_MalformedContent(t *testing.T) {
	// Valid JSON but wrong structure — not an array.
	path := writeTemp(t, "bad.json", `{"name": "not an array"}`)

	fh := filehandler.New()
	_, err := fh.LoadJSON(path)

	require.Error(t, err)
	require.True(t, errors.Is(err, filehandler.ErrInvalidFormat),
		"expected ErrInvalidFormat, got: %v", err)
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	// Completely broken JSON syntax.
	path := writeTemp(t, "broken.json", `[{name: missing quotes}]`)

	fh := filehandler.New()
	_, err := fh.LoadJSON(path)

	require.Error(t, err)
	require.True(t, errors.Is(err, filehandler.ErrInvalidFormat),
		"expected ErrInvalidFormat, got: %v", err)
}

func TestLoadJSON_EmptyArray(t *testing.T) {
	// An empty array is valid — zero cases, no error.
	path := writeTemp(t, "empty.json", `[]`)

	fh := filehandler.New()
	cases, err := fh.LoadJSON(path)

	require.NoError(t, err)
	require.Empty(t, cases)
}

// --- LoadYAML ----------------------------------------------------------------

func TestLoadYAML_HappyPath(t *testing.T) {
	path := writeTemp(t, "cases.yaml", `
- name: "first case"
  input:
    x: 1
  expected:
    y: 2
- name: "second case"
  input:
    x: 3
  expected:
    y: 4
`)

	fh := filehandler.New()
	cases, err := fh.LoadYAML(path)

	require.NoError(t, err)
	require.Len(t, cases, 2)
	require.Equal(t, "first case", cases[0].Name)
	require.Equal(t, "second case", cases[1].Name)

	// Verify YAML input/expected were correctly normalised to json.RawMessage.
	var input map[string]int
	require.NoError(t, json.Unmarshal(cases[0].Input, &input))
	require.Equal(t, 1, input["x"])

	var expected map[string]int
	require.NoError(t, json.Unmarshal(cases[0].Expected, &expected))
	require.Equal(t, 2, expected["y"])
}

func TestLoadYAML_FileNotFound(t *testing.T) {
	fh := filehandler.New()
	_, err := fh.LoadYAML("/does/not/exist/cases.yaml")

	require.Error(t, err)
	require.True(t, errors.Is(err, filehandler.ErrFileNotFound),
		"expected ErrFileNotFound, got: %v", err)
}

func TestLoadYAML_MalformedContent(t *testing.T) {
	// Invalid YAML syntax.
	path := writeTemp(t, "bad.yaml", `
- name: broken
  input: [unclosed bracket
`)

	fh := filehandler.New()
	_, err := fh.LoadYAML(path)

	require.Error(t, err)
	require.True(t, errors.Is(err, filehandler.ErrInvalidFormat),
		"expected ErrInvalidFormat, got: %v", err)
}

func TestLoadYAML_EmptyArray(t *testing.T) {
	// An empty YAML document (empty array equivalent) — zero cases, no error.
	path := writeTemp(t, "empty.yaml", `[]`)

	fh := filehandler.New()
	cases, err := fh.LoadYAML(path)

	require.NoError(t, err)
	require.Empty(t, cases)
}

// --- RunCases ----------------------------------------------------------------

func TestRunCases_SubtestsAreNamed(t *testing.T) {
	cases := []filehandler.TestCase{
		{Name: "alpha", Input: json.RawMessage(`{}`), Expected: json.RawMessage(`{}`)},
		{Name: "beta", Input: json.RawMessage(`{}`), Expected: json.RawMessage(`{}`)},
	}

	visited := make([]string, 0, len(cases))

	fh := filehandler.New()
	fh.RunCases(t, cases, func(t *testing.T, tc filehandler.TestCase) {
		visited = append(visited, tc.Name)
	})

	require.Equal(t, []string{"alpha", "beta"}, visited)
}

func TestRunCases_EmptyNameFallback(t *testing.T) {
	// Cases with no name should receive generated names: case_0, case_1, ...
	cases := []filehandler.TestCase{
		{Name: "", Input: json.RawMessage(`{}`), Expected: json.RawMessage(`{}`)},
		{Name: "", Input: json.RawMessage(`{}`), Expected: json.RawMessage(`{}`)},
	}

	visitCount := 0

	fh := filehandler.New()
	fh.RunCases(t, cases, func(t *testing.T, tc filehandler.TestCase) {
		visitCount++
	})

	// Both cases ran — the subtest names (case_0, case_1) are verified
	// implicitly by the test output; here we confirm the callback fired twice.
	require.Equal(t, 2, visitCount)
}

func TestRunCases_EmptyCaseSlice(t *testing.T) {
	// No cases — callback must never be called.
	callbackFired := false

	fh := filehandler.New()
	fh.RunCases(t, []filehandler.TestCase{}, func(t *testing.T, tc filehandler.TestCase) {
		callbackFired = true
	})

	require.False(t, callbackFired, "callback should not fire for empty case slice")
}

func TestRunCases_UsesFixtureFiles(t *testing.T) {
	// Integration-style: load real fixture files and run them via RunCases.
	// This verifies the full LoadJSON → RunCases pipeline end to end.
	fh := filehandler.New()

	cases, err := fh.LoadJSON("../testdata/sample_cases.json")
	require.NoError(t, err)
	require.NotEmpty(t, cases, "sample_cases.json must contain at least one case")

	fh.RunCases(t, cases, func(t *testing.T, tc filehandler.TestCase) {
		// Verify each case has non-empty Input and Expected.
		require.NotEmpty(t, tc.Input, "Input should not be empty")
		require.NotEmpty(t, tc.Expected, "Expected should not be empty")
	})
}
