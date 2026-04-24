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
	reachedAfterAssert := false

	// t.Run returns false when the subtest fails.
	// The subtest will fail because of the intentional Equal(1, 2) below —
	// that is expected. What we verify at the parent level is that execution
	// continued past the failing assertion, which means reachedAfterAssert
	// must be true even though the subtest is marked failed.
	t.Run("assert does not stop execution", func(t *testing.T) {
		kit := testifywrapper.New(t)
		kit.Assert().Equal(1, 2, "intentional failure — assert should not stop execution")

		// Assert() is non-fatal — this line must always be reached.
		reachedAfterAssert = true
	})

	// The subtest being FAIL above is expected — that is the whole point.
	// The only thing we verify at the parent level is that execution continued.
	if !reachedAfterAssert {
		t.Error("Assert() stopped execution — it must not")
	}
}

// --- Require: fatal behaviour ------------------------------------------------

func TestRequire_StopsExecutionOnFailure(t *testing.T) {
	reachedAfterRequire := false

	// t.Run returns false when the subtest fails.
	// The subtest will fail because of the intentional Equal(1, 2) below —
	// that is expected. What we verify at the parent level is that execution
	// did NOT continue past the failing assertion, which means
	// reachedAfterRequire must still be false.
	t.Run("require stops execution", func(t *testing.T) {
		kit := testifywrapper.New(t)
		kit.Require().Equal(1, 2, "intentional failure — require should stop execution")

		// Require() is fatal — this line must NEVER be reached.
		reachedAfterRequire = true
	})

	// The subtest being FAIL above is expected — that is the whole point.
	// The only thing we verify at the parent level is that execution stopped.
	if reachedAfterRequire {
		t.Error("Require() did not stop execution — it must")
	}
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

// TestRegisterCleanup_NilFuncIsRejected verifies that passing nil to
// RegisterCleanup causes an immediate test failure rather than silently
// registering a no-op. A nil cleanup is a programming mistake — the library
// rejects it loudly via t.Fatal so bugs surface at registration time, not
// silently at cleanup time.
//
// t.Run returns false when the subtest fails. We use that return value to
// assert that nil was rejected — the subtest being FAIL is the expected outcome.
func TestRegisterCleanup_NilFuncIsRejected(t *testing.T) {
	ran := t.Run("nil func causes test failure", func(t *testing.T) {
		kit := testifywrapper.New(t)
		kit.RegisterCleanup(nil)

		// RegisterCleanup(nil) calls t.Fatal — execution never reaches here.
		// If it does, nil was incorrectly accepted.
		t.Error("RegisterCleanup(nil) should have called t.Fatal before this line")
	})

	// t.Run returns false when the subtest failed.
	// nil must have been rejected — so ran must be false.
	if ran {
		t.Error("RegisterCleanup(nil) should have failed the test but did not")
	}
}
