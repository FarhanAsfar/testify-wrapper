package filehandler_test

import (
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/filehandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_RunCases(t *testing.T) {
	h := filehandler.New(t)

	// Test JSON
	err := h.RunCases("../testdata/sample_cases.json", func(t *testing.T, tc filehandler.TestCase) {
		assert.NotEmpty(t, tc.Name)
		assert.NotNil(t, tc.Input)
		assert.NotNil(t, tc.Expected)

		a := tc.Input["a"].(float64)
		b := tc.Input["b"].(float64)
		expected := tc.Expected["result"].(float64)

		assert.Equal(t, expected, a+b)
	})
	require.NoError(t, err)

	// Test YAML
	err = h.RunCases("../testdata/sample_cases.yaml", func(t *testing.T, tc filehandler.TestCase) {
		assert.NotEmpty(t, tc.Name)
		assert.NotNil(t, tc.Input)
		assert.NotNil(t, tc.Expected)

		a := tc.Input["a"].(int)
		b := tc.Input["b"].(int)
		expected := tc.Expected["result"].(int)

		assert.Equal(t, expected, a+b)
	})
	require.NoError(t, err)
}

func TestHandler_LoadCases_Errors(t *testing.T) {
	h := filehandler.New(t)

	_, err := h.LoadCases("nonexistent.json")
	assert.Error(t, err)

	_, err = h.LoadCases("../go.mod") // Unsupported extension
	assert.Error(t, err)
}
