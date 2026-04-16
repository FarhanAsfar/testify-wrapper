// Package examples_test demonstrates full usage of testifyWrapper for new team members.
// This file is both a runnable test and the primary onboarding document.
// Read it top to bottom on day one.
//
// It covers three patterns you will use in every W3 Engineers Go project:
//
//  1. Simple test  — kit := testifyWrapper.New(t), use Assert() and Require()
//  2. Suite test   — struct-based suite with lifecycle hooks via suite.Run()
//  3. File-driven  — load fixture cases from JSON/YAML via kit.LoadJSON/LoadYAML
//
// Depends on:
//   - encoding/json (stdlib)
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper
//   - github.com/FarhanAsfar/testify-wrapper/suite
//
// Used by:
//   - go test ./examples/...
package examples_test

import (
	"encoding/json"
	"testing"

	testifywrapper "github.com/FarhanAsfar/testify-wrapper"
	"github.com/FarhanAsfar/testify-wrapper/suite"
)

// =============================================================================
// The domain: a simple in-memory Calculator.
//
// We use this across all three patterns so the focus stays on the library,
// not on the business logic. In a real project this might be a service,
// a repository, or an HTTP client — the testing patterns are identical.
// =============================================================================

// Calculator performs basic arithmetic and records its operation history.
type Calculator struct {
	history []string
}

// Add returns the sum of a and b.
func (c *Calculator) Add(a, b int) int {
	c.history = append(c.history, "add")
	return a + b
}

// Subtract returns a minus b.
func (c *Calculator) Subtract(a, b int) int {
	c.history = append(c.history, "subtract")
	return a - b
}

// Reset clears the operation history.
func (c *Calculator) Reset() {
	c.history = nil
}

// =============================================================================
// Pattern 1 — Simple test with testifyWrapper.New(t)
//
// This is the most common pattern. One line at the top gives you Assert(),
// Require(), LoadJSON/LoadYAML, RunCases, and RegisterCleanup — all bound
// to the current test's *testing.T.
// =============================================================================

func TestCalculator_Add_Simple(t *testing.T) {
	// kit is your single entry point for everything testifyWrapper provides.
	kit := testifywrapper.New(t)

	calc := &Calculator{}

	// Require() stops the test immediately on failure.
	// Use it when there is no point continuing if this check fails.
	kit.Require().NotNil(calc, "calculator must be initialised before use")

	result := calc.Add(2, 3)

	// Assert() records the failure but lets the test keep running.
	// Use it when you want to collect multiple failures in one run.
	kit.Assert().Equal(5, result, "2 + 3 should equal 5")
	kit.Assert().Len(calc.history, 1, "one operation should be recorded")
}

func TestCalculator_Subtract_Simple(t *testing.T) {
	kit := testifywrapper.New(t)
	calc := &Calculator{}

	result := calc.Subtract(10, 4)

	kit.Assert().Equal(6, result, "10 - 4 should equal 6")
}

func TestCalculator_RegisterCleanup(t *testing.T) {
	kit := testifywrapper.New(t)
	calc := &Calculator{}

	// RegisterCleanup is the right way to schedule teardown inside a test.
	// It runs even if the test calls t.Fatal or panics — defer does not
	// guarantee this in all cases.
	kit.RegisterCleanup(func() {
		calc.Reset()
		// In a real test: close a DB connection, stop a server, etc.
	})

	calc.Add(1, 2)
	kit.Assert().Len(calc.history, 1, "one operation should be in history")
}

// =============================================================================
// Pattern 2 — Struct-based suite with lifecycle hooks
//
// Use this pattern when tests share expensive setup — a database connection,
// a running server, a seeded cache. SetupSuite runs once for the whole suite.
// SetupTest / TearDownTest wrap each individual Test* method.
// Shutdown is the last thing that runs, guaranteed — even on failure.
//
// Lifecycle order:
//   SetupSuite
//     SetupTest → TestXxx → TearDownTest   (repeated per Test* method)
//   TearDownSuite
//   Shutdown
// =============================================================================

// CalculatorSuite groups all Calculator tests that share initialisation.
// Embed suite.BaseSuite to get no-op defaults — override only what you need.
type CalculatorSuite struct {
	suite.BaseSuite

	// calc is the shared resource initialised once in SetupSuite.
	// In a real project this might be *sql.DB, *http.Client, etc.
	calc *Calculator
}

// SetupSuite runs once before any Test* method.
// Initialise shared, expensive resources here.
func (s *CalculatorSuite) SetupSuite() {
	s.calc = &Calculator{}
}

// SetupTest runs before each individual Test* method.
// Reset state here so tests never bleed into each other.
func (s *CalculatorSuite) SetupTest() {
	s.calc.Reset()
}

// TearDownTest runs after each individual Test* method.
// Light per-test cleanup goes here.
func (s *CalculatorSuite) TearDownTest() {
	// Nothing needed for Calculator.
	// In a real suite: roll back a DB transaction, clear mock recorded calls.
}

// Shutdown runs once after all Test* methods and after TearDownSuite.
// Release long-lived resources here — this is the guaranteed-final hook.
func (s *CalculatorSuite) Shutdown() {
	// In a real suite: s.db.Close(), s.server.Stop(), etc.
	s.calc = nil
}

// TestAdd verifies addition using the shared calculator.
//
// Note: suite Test* methods do not receive *testing.T directly in this MVP.
// Use panic() as a hard stop — the Go test runner catches it as a test failure.
// Storing *testing.T on the suite in SetupTest is the recommended post-MVP pattern.
func (s *CalculatorSuite) TestAdd() {
	result := s.calc.Add(3, 4)
	if result != 7 {
		panic("TestAdd: expected 3 + 4 = 7")
	}
}

// TestSubtract verifies subtraction using the shared calculator.
func (s *CalculatorSuite) TestSubtract() {
	result := s.calc.Subtract(10, 3)
	if result != 7 {
		panic("TestSubtract: expected 10 - 3 = 7")
	}
}

// TestHistoryIsResetBetweenTests verifies that SetupTest's Reset() call
// means each test starts with a clean history — no bleed between tests.
func (s *CalculatorSuite) TestHistoryIsResetBetweenTests() {
	if len(s.calc.history) != 0 {
		panic("TestHistoryIsResetBetweenTests: history was not reset before this test")
	}

	s.calc.Add(1, 1)

	if len(s.calc.history) != 1 {
		panic("TestHistoryIsResetBetweenTests: expected exactly one history entry after Add")
	}
}

// TestCalculatorSuite is the standard Go test function that hands off to
// suite.Run. This is the only function the Go test runner calls directly.
func TestCalculatorSuite(t *testing.T) {
	suite.Run(t, &CalculatorSuite{})
}

// =============================================================================
// Pattern 3 — File-driven tests with LoadJSON / LoadYAML
//
// Use this pattern when the same logic needs to run against many input/output
// pairs. Define your cases in a JSON or YAML fixture file — no boilerplate
// in the test file itself. testifyWrapper handles the loop and subtest wiring.
//
// Fixture file schema (JSON or YAML):
//
//	[
//	  {
//	    "name":     "descriptive name — becomes the subtest label",
//	    "input":    { ...your input fields... },
//	    "expected": { ...your expected output fields... }
//	  }
//	]
// =============================================================================

// addInput matches the "input" field schema in our fixture files.
type addInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

// addExpected matches the "expected" field schema in our fixture files.
type addExpected struct {
	Result int `json:"result"`
}

func TestCalculator_Add_FileDriven_JSON(t *testing.T) {
	kit := testifywrapper.New(t)
	calc := &Calculator{}

	// Load cases from the JSON fixture file.
	// If the file is missing or malformed, the test fails loudly here —
	// never silently with zero cases run.
	cases, err := kit.LoadJSON("../testdata/sample_cases.json")
	kit.Require().NoError(err, "fixture file must load without error")
	kit.Require().NotEmpty(cases, "fixture file must contain at least one case")

	// RunCases wires each case as a named subtest.
	// testifyWrapper owns the loop — you own the assertion logic inside fn.
	kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
		// Always create a new kit bound to the subtest's own *testing.T.
		// Never reuse the parent kit inside a subtest.
		subKit := testifywrapper.New(t)

		var input addInput
		subKit.Require().NoError(
			json.Unmarshal(tc.Input, &input),
			"input field must unmarshal into addInput",
		)

		var expected addExpected
		subKit.Require().NoError(
			json.Unmarshal(tc.Expected, &expected),
			"expected field must unmarshal into addExpected",
		)

		result := calc.Add(input.A, input.B)

		subKit.Assert().Equal(
			expected.Result,
			result,
			"Add(%d, %d): expected %d, got %d", input.A, input.B, expected.Result, result,
		)
	})
}

func TestCalculator_Add_FileDriven_YAML(t *testing.T) {
	kit := testifywrapper.New(t)
	calc := &Calculator{}

	// Identical pattern to the JSON version above.
	// Input and Expected are always json.RawMessage regardless of source format,
	// so the subtest body is exactly the same — only the loader changes.
	cases, err := kit.LoadYAML("../testdata/sample_cases.yaml")
	kit.Require().NoError(err, "fixture file must load without error")
	kit.Require().NotEmpty(cases, "fixture file must contain at least one case")

	kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
		subKit := testifywrapper.New(t)

		var input addInput
		subKit.Require().NoError(
			json.Unmarshal(tc.Input, &input),
			"input field must unmarshal into addInput",
		)

		var expected addExpected
		subKit.Require().NoError(
			json.Unmarshal(tc.Expected, &expected),
			"expected field must unmarshal into addExpected",
		)

		result := calc.Add(input.A, input.B)

		subKit.Assert().Equal(
			expected.Result,
			result,
			"Add(%d, %d): expected %d, got %d", input.A, input.B, expected.Result, result,
		)
	})
}
