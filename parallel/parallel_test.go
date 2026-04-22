// Package parallel_test verifies the behaviour of parallel.Configure,
// parallel.Reset, and parallel.IsEnabled.
//
// Depends on:
//   - parallel (this module)
//   - runtime (stdlib)
//   - testing (stdlib)
//
// Used by:
//   - go test ./parallel/...
package parallel_test

import (
	"runtime"
	"testing"

	"github.com/FarhanAsfar/testify-wrapper/parallel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetAfter ensures every test case leaves the process in a clean state
// regardless of what Configure did. Without this, GOMAXPROCS changes made
// by one test would bleed into the next.
func resetAfter(t *testing.T) {
	t.Helper()
	t.Cleanup(parallel.Reset)
}

// TestConfigure_Disabled verifies that Configure with Enabled:false is a
// complete no-op — GOMAXPROCS is unchanged and IsEnabled returns false.
func TestConfigure_Disabled(t *testing.T) {
	resetAfter(t)

	before := runtime.GOMAXPROCS(0)

	parallel.Configure(parallel.Config{
		Enabled:  false,
		MaxProcs: 4,
	})

	// IsEnabled must be false.
	assert.False(t, parallel.IsEnabled(),
		"IsEnabled should be false when Enabled:false")

	// GOMAXPROCS must not have changed — Enabled:false means we never
	// called maxprocs.Set, so the value should be identical.
	assert.Equal(t, before, runtime.GOMAXPROCS(0),
		"GOMAXPROCS should be unchanged when parallel is disabled")
}

// TestConfigure_Enabled_AutoMaxProcs verifies that Configure with Enabled:true
// and MaxProcs:0 activates parallel mode and lets automaxprocs decide
// GOMAXPROCS. We cannot assert an exact GOMAXPROCS value here because
// automaxprocs reads CPU quota from the environment, but we can assert that
// the value is at least 1 (a valid, usable runtime).
func TestConfigure_Enabled_AutoMaxProcs(t *testing.T) {
	resetAfter(t)

	parallel.Configure(parallel.Config{
		Enabled:  true,
		MaxProcs: 0, // automaxprocs decides
	})

	require.True(t, parallel.IsEnabled(),
		"IsEnabled should be true after Configure with Enabled:true")

	got := runtime.GOMAXPROCS(0)
	assert.GreaterOrEqual(t, got, 1,
		"GOMAXPROCS should be at least 1 after automaxprocs configuration")
}

// TestConfigure_Enabled_ExplicitMaxProcs verifies that Configure with
// MaxProcs > 0 sets GOMAXPROCS to at least the requested floor value.
// automaxprocs.Min guarantees the floor — the actual value may be higher
// if the CPU quota allows, so we use GreaterOrEqual, not Equal.
func TestConfigure_Enabled_ExplicitMaxProcs(t *testing.T) {
	resetAfter(t)

	parallel.Configure(parallel.Config{
		Enabled:  true,
		MaxProcs: 2,
	})

	require.True(t, parallel.IsEnabled(),
		"IsEnabled should be true after Configure with Enabled:true")

	got := runtime.GOMAXPROCS(0)
	assert.GreaterOrEqual(t, got, 2,
		"GOMAXPROCS should be at least the requested MaxProcs floor")
}

// TestReset_RestoresGOMAXPROCS verifies that Reset undoes the GOMAXPROCS
// change made by Configure and sets IsEnabled back to false.
func TestReset_RestoresGOMAXPROCS(t *testing.T) {
	// Do not use resetAfter here — we are testing Reset explicitly.
	// We do however need to clean up if the test itself fails mid-way,
	// so we register a safety-net cleanup that is harmless if Reset
	// has already been called (Reset is idempotent).
	t.Cleanup(parallel.Reset)

	before := runtime.GOMAXPROCS(0)

	parallel.Configure(parallel.Config{
		Enabled:  true,
		MaxProcs: 0,
	})

	require.True(t, parallel.IsEnabled(),
		"IsEnabled should be true after Configure")

	parallel.Reset()

	assert.False(t, parallel.IsEnabled(),
		"IsEnabled should be false after Reset")

	assert.Equal(t, before, runtime.GOMAXPROCS(0),
		"GOMAXPROCS should be restored to its pre-Configure value after Reset")
}

// TestReset_IsIdempotent verifies that calling Reset multiple times does not
// panic or produce unexpected state. The double-undo guard in parallel.go
// makes this safe — this test confirms that guarantee holds.
func TestReset_IsIdempotent(t *testing.T) {
	parallel.Configure(parallel.Config{Enabled: true, MaxProcs: 0})

	// Calling Reset twice must not panic.
	require.NotPanics(t, func() {
		parallel.Reset()
		parallel.Reset()
	}, "Reset should be safe to call multiple times")

	assert.False(t, parallel.IsEnabled(),
		"IsEnabled should be false after repeated Reset calls")
}

// TestConfigure_CalledTwice verifies that calling Configure a second time
// correctly undoes the first call before applying the new settings.
// This covers the case where a consumer calls Configure twice in TestMain
// (e.g. during test development or reconfiguration).
func TestConfigure_CalledTwice(t *testing.T) {
	resetAfter(t)

	// First Configure — enable with an explicit floor.
	parallel.Configure(parallel.Config{Enabled: true, MaxProcs: 2})
	require.True(t, parallel.IsEnabled())

	// Second Configure — disable entirely.
	parallel.Configure(parallel.Config{Enabled: false})

	assert.False(t, parallel.IsEnabled(),
		"IsEnabled should reflect the most recent Configure call")
}
