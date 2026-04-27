# testifyWrapper

A consistent, opinionated Go testing harness, built on top of [testify](https://github.com/stretchr/testify).

One import. One `New(t)` call. We get assertions, lifecycle hooks,
resource cleanup, and file-driven test execution — all wired together.

---

## Install

```bash
go get github.com/FarhanAsfar/testify-wrapper
```

---

## Quick Start

### Simple test

```go
import testifywrapper "github.com/FarhanAsfar/testify-wrapper"

func TestAdd(t *testing.T) {
    kit := testifywrapper.New(t)

    result := Add(2, 3)

    // Assert() records failure but keeps running — collect all failures at once.
    kit.Assert().Equal(5, result)

    // Require() stops immediately — use when continuing makes no sense.
    kit.Require().NoError(someErr)
}
```

---

### Struct-based suite with lifecycle hooks

Embed `BaseSuite` for no-op defaults and override only the hooks you need.
Access the current subtest's `*testing.T` inside any method via `s.T()`.

```go
import (
    testifywrapper "github.com/FarhanAsfar/testify-wrapper"
    "github.com/FarhanAsfar/testify-wrapper/suite"
)

type MyServiceSuite struct {
    suite.BaseSuite   // provides s.T() and no-op lifecycle defaults
    db *sql.DB
}

// SetupSuite runs once before any Test* method.
// Note: s.T() is nil here — it is only bound during Test* method execution.
func (s *MyServiceSuite) SetupSuite() {
    s.db = connectTestDB()
}

// SetupTest runs before each Test* method.
// s.T() is valid here and points to the current subtest's *testing.T.
func (s *MyServiceSuite) SetupTest() {
    s.db.Exec("DELETE FROM orders") // reset state between tests
}

// Shutdown is the guaranteed-final hook — release long-lived resources here.
func (s *MyServiceSuite) Shutdown() {
    s.db.Close()
}

// Test* methods use s.T() to get the subtest-scoped *testing.T.
// Wrap it with testifyWrapper.New() to get the full assertion API.
func (s *MyServiceSuite) TestCreateOrder() {
    kit := testifywrapper.New(s.T())

    order, err := CreateOrder(s.db, "item-1")

    kit.Require().NoError(err, "CreateOrder must not return an error")
    kit.Assert().Equal("item-1", order.Item)
}

func (s *MyServiceSuite) TestListOrders() {
    kit := testifywrapper.New(s.T())

    orders, err := ListOrders(s.db)

    kit.Require().NoError(err)
    kit.Assert().NotEmpty(orders)
}

// TestMyServiceSuite is the only function the Go test runner calls directly.
func TestMyServiceSuite(t *testing.T) {
    suite.Run(t, &MyServiceSuite{})
}
```

**Lifecycle order:**
```
SetupSuite                                      ← s.T() is nil here
  SetupTest → TestXxx → TearDownTest            ← s.T() is valid here (once per Test*)
TearDownSuite                                   ← s.T() is nil here
Shutdown                                        ← s.T() is nil here
```

**`s.T()` availability:**

| Hook            | `s.T()` valid? |
|-----------------|----------------|
| `SetupSuite`    | ❌ nil          |
| `SetupTest`     | ✅ yes          |
| `TestXxx`       | ✅ yes          |
| `TearDownTest`  | ✅ yes          |
| `TearDownSuite` | ❌ nil          |
| `Shutdown`      | ❌ nil          |

---

### File-driven test

```go
func TestAdd_FileDriven(t *testing.T) {
    kit := testifywrapper.New(t)

    cases, err := kit.LoadJSON("testdata/cases.json")
    kit.Require().NoError(err)

    kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
        // Always create a new kit bound to the subtest's own *testing.T.
        subKit := testifywrapper.New(t)

        var input struct{ A, B int }
        subKit.Require().NoError(json.Unmarshal(tc.Input, &input))

        var expected struct{ Result int }
        subKit.Require().NoError(json.Unmarshal(tc.Expected, &expected))

        subKit.Assert().Equal(expected.Result, Add(input.A, input.B))
    })
}
```

---

## Packages

| Package                 | Purpose                                                              |
|-------------------------|----------------------------------------------------------------------|
| `testifyWrapper` (root) | Entry point — `New(t)` returns an `Instance` with everything wired  |
| `assert`                | Non-fatal assertions — test keeps running after failure              |
| `require`               | Fatal assertions — test stops immediately after failure              |
| `suite`                 | Struct-based suite runner with ordered lifecycle hooks and `s.T()`   |
| `filehandler`           | JSON/YAML fixture loader for data-driven tests                       |
| `parallel`              | Parallel suite configuration — `Configure` and `Reset`              |
| `internal/hooks`        | Internal cleanup registry (not for direct use)                       |

---

## Fixture File Format

Both JSON and YAML are supported. Every fixture file must be a top-level
array where each entry has three fields:

| Field      | Type   | Description                                                  |
|------------|--------|--------------------------------------------------------------|
| `name`     | string | Becomes the subtest label in `go test` output                |
| `input`    | object | Your input — unmarshal into your own struct in the callback  |
| `expected` | object | Your expected output — unmarshal into your own struct        |

**JSON example** (`testdata/cases.json`):
```json
[
  {
    "name": "add two positive numbers",
    "input":    { "a": 2, "b": 3 },
    "expected": { "result": 5 }
  },
  {
    "name": "add zero to a number",
    "input":    { "a": 7, "b": 0 },
    "expected": { "result": 7 }
  }
]
```

**YAML example** (`testdata/cases.yaml`):
```yaml
- name: "add two positive numbers"
  input:
    a: 2
    b: 3
  expected:
    result: 5

- name: "add zero to a number"
  input:
    a: 7
    b: 0
  expected:
    result: 7
```

`input` and `expected` are always delivered as `json.RawMessage` inside
the callback — regardless of whether the fixture was JSON or YAML.
Unmarshal them into your own concrete types.

---

## Error Handling

`LoadJSON` and `LoadYAML` return typed sentinel errors.
Use `errors.Is()` for specific handling:

```go
cases, err := kit.LoadJSON("testdata/cases.json")
if errors.Is(err, filehandler.ErrFileNotFound) {
    // fixture file missing — check path
}
if errors.Is(err, filehandler.ErrInvalidFormat) {
    // fixture file exists but content is malformed
}
```

---

## Key Behaviours to Know

**Always use `RegisterCleanup` over `defer` for teardown inside tests.**
`t.Cleanup` runs even when the test calls `t.Fatal` or panics.

```go
kit.RegisterCleanup(func() {
    server.Stop()
})
```

**Never share an `Instance` across tests or goroutines.**
Each `Instance` is bound to one `*testing.T`. Create a new one per test function.

**Always create a new `kit` inside `RunCases` callbacks.**
The callback receives a subtest's `*testing.T` — bind a fresh kit to it:

```go
kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
    subKit := testifywrapper.New(t)  // ← always do this
    subKit.Assert().Equal(...)
})
```

**`s.T()` is only valid during `SetupTest`, `TestXxx`, and `TearDownTest`.**
It is nil during `SetupSuite`, `TearDownSuite`, and `Shutdown` because those
hooks run outside the subtest scope. If you need to log or assert during
suite-level hooks, store a reference to the parent `*testing.T` manually
in your suite struct when `suite.Run(t, s)` is called.

---

## Parallel Testing

Parallel mode allows separate suites within the same package to run
concurrently. Configure it once from `TestMain`, before any test executes.

```go
func TestMain(m *testing.M) {
    testifywrapper.ConfigureParallel(testifywrapper.ParallelConfig{
        Enabled:  true,
        MaxProcs: 0, // let automaxprocs decide — recommended for CI and containers
    })
    defer testifywrapper.ResetParallel()
    os.Exit(m.Run())
}
```

**What `MaxProcs` does:**

| Value        | Behaviour                                                                 |
|--------------|---------------------------------------------------------------------------|
| `0` (default)| automaxprocs reads the Linux cgroup CPU quota, falls back to `NumCPU()`  |
| `> 0`        | GOMAXPROCS is set to at least this value (used as a floor)               |

`MaxProcs: 0` is the correct default for CI and containerised environments.
Use an explicit value only when you need to constrain or guarantee a minimum
thread count on a known machine.

**What runs in parallel:**
When enabled, each `suite.Run(t, s)` call marks its parent test function
as parallel. This means `TestUserSuite` and `TestOrderSuite` run concurrently
with each other. The `Test*` methods within a single suite always run
sequentially relative to each other.

**When NOT to enable parallel mode:**

Do not enable parallel mode if any of the following apply to your suite:

- **Shared global state** — environment variables, package-level variables,
  or global caches written by one suite will race with reads from another.
- **Non-isolated database fixtures** — if two suites truncate and seed the
  same tables, they will corrupt each other's data mid-run.
- **Port or file lock contention** — two suites starting servers on the same
  port or writing to the same file path will fail non-deterministically.
- **Execution order dependencies** — if Suite B assumes Suite A has already
  run and left data behind, parallel mode will break that assumption silently.

The safe rule: each suite must be fully self-contained. It must set up
everything it needs in `SetupSuite`, tear it all down in `Shutdown`, and
never read or write anything shared with another suite.

---

## Contributing

**Branch naming:**
- `feat/<topic>` for new functionality
- `fix/<topic>` for bug fixes
- `chore/<topic>` for maintenance (docs, deps, refactor)

**Before opening a PR:**
- `go test ./...` must pass
- `go vet ./...` must pass with zero warnings
- Every new exported symbol must have a GoDoc comment
- Every new feature must include tests — no exceptions
- Update `CHANGELOG.md` under the appropriate version section

**Adding custom assertions:**
Domain-specific assertions (e.g. `AssertValidUUID`) belong in
`assert/assert.go` as methods on `Assertions`. Mirror the same method
in `require/require.go` for the fatal variant.
