// Package examples_test demonstrates full usage of testifyWrapper
// This file is both a runnable test and the primary onboarding document.
// Read it top to bottom on day one.
//
//  1. Simple test    — kit := testifyWrapper.New(t), use Assert() and Require()
//  2. Suite test     — struct-based suite with lifecycle hooks via suite.Run()
//  3. File-driven    — load fixture cases from JSON/YAML via kit.LoadJSON/LoadYAML
//  4. Parallel suite — multiple suites running concurrently via ConfigureParallel()
//
// Depends on:
//   - encoding/json (stdlib)
//   - os (stdlib)
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper
//   - github.com/FarhanAsfar/testify-wrapper/suite
//
// Used by:
//   - go test ./examples/...
package examples_test

import (
	"encoding/json"
	"os"
	"testing"

	testifywrapper "github.com/FarhanAsfar/testify-wrapper"
	"github.com/FarhanAsfar/testify-wrapper/suite"
)

// =============================================================================
// The domain: a simple in-memory Calculator.
//
// We use this across all four patterns so the focus stays on the library,
// not on the business logic. In a real project this might be a service,
// a repository, or an HTTP client — the testing patterns are identical.
// =============================================================================

// Calculator performs basic arithmetic and records its operation history.
type Calculator struct {
	// history records every operation performed, in order.
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
// Pattern 4 — Parallel suites via TestMain
//
// TestMain is the process-level entry point for the test binary.
// It runs once before any test function executes — the right place to
// configure process-wide settings like parallelism.
//
// When ConfigureParallel is called with Enabled:true, every suite.Run()
// call in this package will invoke t.Parallel() on its parent test function.
// This allows TestCalculatorSuite and TestAuditSuite (below) to run
// concurrently with each other rather than sequentially.
//
// MaxProcs:0 tells automaxprocs to decide the GOMAXPROCS value — reading
// the Linux cgroup CPU quota if available, falling back to runtime.NumCPU().
// This is the recommended setting for CI and containerised environments.
//
// ResetParallel restores GOMAXPROCS to its original value after the run.
// =============================================================================

func TestMain(m *testing.M) {
	testifywrapper.ConfigureParallel(testifywrapper.ParallelConfig{
		Enabled:  true,
		MaxProcs: 0, // let automaxprocs decide — recommended for CI
	})
	defer testifywrapper.ResetParallel()

	os.Exit(m.Run())
}

// =============================================================================
// Pattern 1 — Simple test with testifyWrapper.New(t)
//
// This is the most common pattern. One line at the top of your test gives you
// Assert(), Require(), LoadJSON(), LoadYAML(), RunCases(), and RegisterCleanup().
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
	kit.Assert().Len(calc.history, 1, "Add should record exactly one history entry")
}

func TestCalculator_RegisterCleanup(t *testing.T) {
	kit := testifywrapper.New(t)
	calc := &Calculator{}

	calc.Add(1, 1)

	// RegisterCleanup runs after the test ends, even on t.Fatal or panic.
	// Prefer this over defer for all teardown in test code.
	kit.RegisterCleanup(func() {
		calc.Reset()
	})

	kit.Assert().Len(calc.history, 1)
}

// =============================================================================
// Pattern 2 — Suite test with lifecycle hooks
//
// Use suites when multiple tests share expensive setup — a real DB connection,
// an HTTP server, a seeded cache. The suite struct holds the shared resource.
// BaseSuite provides no-op defaults for all five hooks — override only what
// you need.
//
// Lifecycle order (guaranteed):
//   SetupSuite  → once before all Test* methods
//   SetupTest   → before each Test* method
//   TearDownTest → after each Test* method
//   TearDownSuite → once after all Test* methods
//   Shutdown    → last, always — release long-lived resources here
// =============================================================================

// CalculatorSuite demonstrates a suite with a shared Calculator instance.
// SetupSuite initialises the calculator once. SetupTest resets it between
// tests so each Test* method starts from a clean state.
type CalculatorSuite struct {
	suite.BaseSuite
	calc *Calculator
}

func (s *CalculatorSuite) SetupSuite() {
	// Initialise the shared resource once for all Test* methods.
	// In a real suite this might open a DB connection or start a server.
	s.calc = &Calculator{}
}

func (s *CalculatorSuite) SetupTest() {
	// Reset state before each test so methods are fully independent.
	s.calc.Reset()
}

func (s *CalculatorSuite) Shutdown() {
	// Release long-lived resources here — always runs last.
	// In a real suite: s.db.Close(), s.server.Stop(), etc.
	s.calc = nil
}

func (s *CalculatorSuite) TestAdd() {
	kit := testifywrapper.New(s.T())
	result := s.calc.Add(3, 4)
	kit.Assert().Equal(7, result, "3 + 4 should equal 7")
}

func (s *CalculatorSuite) TestSubtract() {
	kit := testifywrapper.New(s.T())
	result := s.calc.Subtract(10, 3)
	kit.Assert().Equal(7, result, "10 - 3 should equal 7")
}

func (s *CalculatorSuite) TestHistoryIsResetBetweenTests() {
	kit := testifywrapper.New(s.T())

	// history must be empty — SetupTest called Reset() before this ran.
	kit.Assert().Empty(s.calc.history, "history must be empty at the start of each test")

	s.calc.Add(1, 1)
	kit.Assert().Len(s.calc.history, 1, "expected exactly one history entry after Add")
}

// TestCalculatorSuite is the entry point for the suite runner.
// suite.Run discovers and executes all Test* methods on CalculatorSuite,
// wiring the lifecycle hooks automatically.
//
// Because TestMain called ConfigureParallel with Enabled:true, suite.Run
// will call t.Parallel() here — meaning TestCalculatorSuite and
// TestAuditSuite run concurrently with each other.
func TestCalculatorSuite(t *testing.T) {
	suite.Run(t, &CalculatorSuite{})
}

// =============================================================================
// Pattern 3 — File-driven tests with LoadJSON / LoadYAML
//
// Use file-driven tests when the same logic needs to run against many
// input/expected pairs. The fixture file is the test case definition —
// your test file only contains the assertion logic.
//
// This separates "what cases to test" from "how to test them".
// Adding a new case means editing a JSON/YAML file, not Go code.
// =============================================================================

// addInput and addExpected are the concrete types we unmarshal
// fixture fields into inside the RunCases callback.
type addInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

type addExpected struct {
	Result int `json:"result"`
}

func TestCalculator_Add_JSON(t *testing.T) {
	kit := testifywrapper.New(t)

	cases, err := kit.LoadJSON("../testdata/sample_cases.json")
	kit.Require().NoError(err, "fixture file must load without error")

	kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
		subKit := testifywrapper.New(t)

		var input addInput
		var expected addExpected

		subKit.Require().NoError(json.Unmarshal(tc.Input, &input))
		subKit.Require().NoError(json.Unmarshal(tc.Expected, &expected))

		calc := &Calculator{}
		result := calc.Add(input.A, input.B)

		subKit.Assert().Equal(expected.Result, result)
	})
}

func TestCalculator_Add_YAML(t *testing.T) {
	kit := testifywrapper.New(t)

	// LoadYAML normalises Input and Expected to json.RawMessage —
	// the RunCases callback is identical to the JSON variant above.
	cases, err := kit.LoadYAML("../testdata/sample_cases.yaml")
	kit.Require().NoError(err, "fixture file must load without error")

	kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
		subKit := testifywrapper.New(t)

		var input addInput
		var expected addExpected

		subKit.Require().NoError(json.Unmarshal(tc.Input, &input))
		subKit.Require().NoError(json.Unmarshal(tc.Expected, &expected))

		calc := &Calculator{}
		result := calc.Add(input.A, input.B)

		subKit.Assert().Equal(expected.Result, result)
	})
}

// =============================================================================
// Pattern 4 — A second suite demonstrating parallel execution
//
// When ConfigureParallel is enabled, TestCalculatorSuite and TestAuditSuite
// run concurrently with each other. This is Level A parallelism — the two
// suite parent functions run in parallel. The Test* methods within each
// suite still run sequentially relative to each other.
//
// Key rule for parallel suites: do not share any mutable state between
// suites. Each suite must be fully self-contained. AuditSuite and
// CalculatorSuite each own their own Calculator instance — they never
// share one.
// =============================================================================

// AuditSuite is a second independent suite. Its only purpose in this
// example is to show that two suites run concurrently when parallel is enabled.
// In a real project this might be your OrderSuite, UserSuite, etc.
type AuditSuite struct {
	suite.BaseSuite
	calc *Calculator
}

func (s *AuditSuite) SetupSuite() {
	s.calc = &Calculator{}
}

func (s *AuditSuite) SetupTest() {
	s.calc.Reset()
}

func (s *AuditSuite) Shutdown() {
	s.calc = nil
}

func (s *AuditSuite) TestOperationIsRecorded() {
	kit := testifywrapper.New(s.T())
	s.calc.Add(1, 1)
	kit.Assert().Len(s.calc.history, 1, "Add should append exactly one entry to history")
}

func (s *AuditSuite) TestHistoryIsEmptyAfterReset() {
	kit := testifywrapper.New(s.T())
	s.calc.Add(5, 5)
	s.calc.Reset()
	kit.Assert().Empty(s.calc.history, "Reset should clear all history entries")
}

// TestAuditSuite is the second suite entry point.
// With parallel enabled, this runs concurrently with TestCalculatorSuite.
func TestAuditSuite(t *testing.T) {
	suite.Run(t, &AuditSuite{})
}
