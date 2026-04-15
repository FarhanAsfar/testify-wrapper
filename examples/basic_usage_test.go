// Package examples_test demonstrates full usage of testifywrapper for new team members.
// This file is both a runnable test and the primary onboarding document.
//
// Depends on:
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper
//   - github.com/FarhanAsfar/testify-wrapper/suite
//   - github.com/FarhanAsfar/testify-wrapper/filehandler
//
// Used by:
//   - go test ./examples/...
package examples_test

import (
	"fmt"
	"testing"

	tw "github.com/FarhanAsfar/testify-wrapper"
	"github.com/FarhanAsfar/testify-wrapper/filehandler"
	"github.com/FarhanAsfar/testify-wrapper/suite"
)

// Calculator is a simple service we want to test.
type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a + b
}

// 1. Basic usage with Assert and Require
func TestBasicUsage(t *testing.T) {
	kit := tw.New(t)
	calc := &Calculator{}

	// Assert: continues test on failure
	kit.Assert().Equal(5, calc.Add(2, 3), "Addition should work")
	kit.Assert().NotNil(calc)

	// Require: stops test on failure
	kit.Require().Equal(10, calc.Add(5, 5))
}

// 2. Lifecycle management with Hooks
func TestCleanupHooks(t *testing.T) {
	kit := tw.New(t)

	fmt.Println("Acquiring resource A")
	kit.Hooks().Register(func() {
		fmt.Println("Cleaning up resource A")
	})

	fmt.Println("Acquiring resource B")
	kit.Hooks().Register(func() {
		fmt.Println("Cleaning up resource B")
	})

	kit.Assert().True(true)
}

// 3. File-driven tests
func TestFileDriven(t *testing.T) {
	kit := tw.New(t)
	calc := &Calculator{}

	err := kit.FileHandler().RunCases("../testdata/sample_cases.json", func(t *testing.T, tc filehandler.TestCase) {
		a := int(tc.Input["a"].(float64))
		b := int(tc.Input["b"].(float64))
		expected := int(tc.Expected["result"].(float64))

		if calc.Add(a, b) != expected {
			t.Errorf("expected %d, got %d", expected, calc.Add(a, b))
		}
	})
	kit.Require().NoError(err)
}

// 4. Struct-based Test Suite
type CalculatorSuite struct {
	suite.BaseSuite
	calc *Calculator
}

func (s *CalculatorSuite) SetupSuite() {
	s.calc = &Calculator{}
}

func (s *CalculatorSuite) TestAdd() {
	kit := tw.New(s.T)
	kit.Assert().Equal(4, s.calc.Add(2, 2))
}

func (s *CalculatorSuite) TestMoreAdd() {
	kit := tw.New(s.T)
	kit.Assert().Equal(0, s.calc.Add(0, 0))
}

func TestCalculatorSuite(t *testing.T) {
	suite.Run(t, new(CalculatorSuite))
}
