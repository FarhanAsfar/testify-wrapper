# Changelog

All notable changes to this project will be documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [v0.1.0] — MVP

### Added

- `testifyWrapper.New(t)` — core `Instance` entry point binding assert,
  require, filehandler, and cleanup registry to a single `*testing.T`
- `Instance.Assert()` — non-fatal assertion wrapper around testify/assert.
  Test continues collecting failures after an assertion fails.
- `Instance.Require()` — fatal assertion wrapper around testify/require.
  Test stops immediately at the point of failure.
- `Instance.LoadJSON(path)` — loads a JSON fixture file into `[]TestCase`
- `Instance.LoadYAML(path)` — loads a YAML fixture file into `[]TestCase`,
  normalising Input and Expected to `json.RawMessage`
- `Instance.RunCases(t, cases, fn)` — runs each `TestCase` as a named
  subtest, wiring the loop so the callback owns only the assertion logic
- `Instance.RegisterCleanup(fn)` — enqueues teardown functions in LIFO
  order via `t.Cleanup`, safe for parallel tests and `t.Fatal`
- `Instance.FileHandler()` — direct access to the underlying `FileHandler`
  for advanced usage
- `testifyWrapper.TestCase` — re-export of `filehandler.TestCase` so
  consumers never need to import `filehandler` directly
- `suite.Suite` interface — five lifecycle hooks: `SetupSuite`,
  `SetupTest`, `TearDownTest`, `TearDownSuite`, `Shutdown`
- `suite.BaseSuite` — no-op implementations of all five hooks for embedding
- `suite.Run(t, suite)` — reflection-based Test* method discovery and
  execution with full lifecycle wiring
- `filehandler.ErrFileNotFound` — typed sentinel error for missing fixture files
- `filehandler.ErrInvalidFormat` — typed sentinel error for malformed fixture files
- `examples/basic_usage_test.go` — runnable onboarding example covering
  all three usage patterns