// Package parallel provides parallelism configuration for testifyWrapper.
//
// Depends on:
//   - go.uber.org/automaxprocs/maxprocs
//   - runtime (stdlib)
//
// Used by:
//   - testifyWrapper.Instance (via kit.ConfigureParallel())
//   - suite.Run (reads IsEnabled to decide whether to call t.Parallel())
//   - Consumer TestMain functions (via parallel.Configure directly)
package parallel

import (
	"runtime"

	"go.uber.org/automaxprocs/maxprocs"
)

// Config holds the parallelism settings for a test run
// It is intended to be set once before any test executes.
type Config struct {
	// Enabled controls whether suite.Run calls t.Parallel()
	Enabled bool

	// MaxProcs sets the minimum number of os threads (GOMAXPROCS) available for parallel execution.
	// If MaxProcs is 0, automaxprocs decides the value automatically reading the linux cgroup CPU quota if available, otherwise falling back to runtime.NumCPU().
	// If MaxProcs > 0, that value is used as the floor pased to maxprocs.Min(). The actual GOMAXPROCS will be at least MaxProcs but automaxprocs may set it higher if the CPU quota allows
	MaxProcs int
}

// undoMaxProcs is the undo function returned by maxprocs.Set.
// Calling it restores GOMAXPROCS to the value it had before Configure ran.
// It is initialised to a no-op so Reset() is always safe to call.
var undoMaxProcs = func() {}

// enabled tracks if parallel mode is active or not.
// read by IsEnabled, written only by configure.
var enabled bool

// Configure applies the given Config to the current process.
//
// It should be called once from TestMain, before any test runs.
// Calling it multiple times is safe — each call resets the previous
// GOMAXPROCS change via the stored undo function before applying the new one.
func Configure(cfg Config) {
	// Undo any previous configure call before applying the new settings
	undoMaxProcs()

	enabled = cfg.Enabled

	if !cfg.Enabled {
		// parallel is off - restore undo to no-op and leave GOMAXPROCS along
		undoMaxProcs = func() {}
		return
	}

	var opts []maxprocs.Option

	if cfg.MaxProcs > 0 {
		// consumer specified a floor
		opts = append(opts, maxprocs.Min(cfg.MaxProcs))
	}

	undo, err := maxprocs.Set(opts...)
	if err != nil {
		runtime.GOMAXPROCS(runtime.NumCPU())
		undoMaxProcs = func() {}
		return
	}
	undoMaxProcs = undo
}

// Reset restores GOMAXPROCS to the value it had before Configure was called.
// It also disables parallel mode.
//
// Call this at the end of TestMain to leave the process in a clean state.
func Reset() {
	undoMaxProcs()
	undoMaxProcs = func() {} // prevent double-undo
	enabled = false
}

// IsEnabled reports whether parallel mode is currently active.
// suite.Run reads this to decide whether to call t.Parallel().
func IsEnabled() bool {
	return enabled
}
