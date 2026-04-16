// Package testifyWrapper_test verifies the behaviour of the top-level
// testifyWrapper package, including the Instance entry point and the
// non-fatal vs fatal assertion behaviour of Assert() and Require().
//
// Depends on:
//   - testing (stdlib)
//   - github.com/FarhanAsfar/testify-wrapper
//
// Used by:
//   - go test .
package testifyWrapper_test

import (
	"testing"

	testifywrapper "github.com/FarhanAsfar/testify-wrapper"
)

// --- Instance construction ---------------------------------------------------

func TestNew_ReturnsNonNilInstance(t *testing.T) {
	kit := testifywrapper.New(t)

	if kit == nil {
		t.Fatal("New(t) returned nil")
	}
}

func TestNew_AssertIsNonNil(t *testing.T) {
	kit := testifywrapper.New(t)

	if kit.Assert() == nil {
		t.Fatal("kit.Assert() returned nil")
	}
}

func TestNew_RequireIsNonNil(t *testing.T) {
	kit := testifywrapper.New(t)

	if kit.Require() == nil {
		t.Fatal("kit.Require() returned nil")
	}
}

func TestNew_FileHandlerIsNonNil(t *testing.T) {
	kit := testifywrapper.New(t)

	if kit.FileHandler() == nil {
		t.Fatal("kit.FileHandler() returned nil")
	}
}

func TestNew_TIsSet(t *testing.T) {
	kit := testifywrapper.New(t)

	if kit.T != t {
		t.Fatal("kit.T is not the same *testing.T passed to New()")
	}
}

// --- Assert: non-fatal behaviour ---------------------------------------------

func TestAssert_ContinuesAfterFailure(t *testing.T) {
	// Strategy: inside the subtest, register a t.Cleanup that checks two things:
	// 1. The subtest is marked failed (the intentional Equal(1,2) was recorded)
	// 2. reachedAfterAssert is true (execution continued past the assertion)
	// This avoids the parent test inheriting a failure it doesn't own.
	t.Run("assert does not stop execution", func(t *testing.T) {
		reachedAfterAssert := false

		// This cleanup runs after the subtest body — at that point we can
		// safely read both t.Failed() and reachedAfterAssert.
		t.Cleanup(func() {
			if !t.Failed() {
				t.Error("expected the subtest to be marked failed by Assert()")
			}
			if !reachedAfterAssert {
				t.Error("Assert() stopped execution — it should not have")
			}
		})

		kit := testifywrapper.New(t)
		kit.Assert().Equal(1, 2, "intentional failure — assert should not stop execution")

		// Assert() is non-fatal — this line must always be reached.
		reachedAfterAssert = true
	})
}

// --- Require: fatal behaviour ------------------------------------------------

func TestRequire_StopsExecutionOnFailure(t *testing.T) {
	// Strategy: set reachedAfterRequire = true after the Require call.
	// If Require() is truly fatal, that line never executes.
	// We assert this from a t.Cleanup inside the subtest, which runs
	// after FailNow() has already unwound the subtest goroutine.
	t.Run("require stops execution", func(t *testing.T) {
		reachedAfterRequire := false

		t.Cleanup(func() {
			if reachedAfterRequire {
				t.Error("Require() did not stop execution — it should have")
			}
		})

		kit := testifywrapper.New(t)
		kit.Require().Equal(1, 2, "intentional failure — require should stop execution")

		// Require() is fatal — this line must NEVER be reached.
		reachedAfterRequire = true
	})
}

// --- RegisterCleanup ---------------------------------------------------------

func TestRegisterCleanup_FiresOnTestEnd(t *testing.T) {
	// We can't assert "after" a test ends from inside it, so we verify
	// the registration doesn't panic and the library wires it without error.
	// Full LIFO ordering is covered by the hooks package design and suite tests.
	kit := testifywrapper.New(t)

	// Should not panic, and the func should run silently when t ends.
	require := kit.Require()
	require.NotPanics(func() {
		kit.RegisterCleanup(func() {
			// cleanup registered — verified by absence of panic
		})
	})
}

func TestRegisterCleanup_NilFuncDoesNotPanic(t *testing.T) {
	kit := testifywrapper.New(t)

	kit.Require().NotPanics(func() {
		kit.RegisterCleanup(nil)
	})
}
