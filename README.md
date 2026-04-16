# testifyWrapper

A consistent, opinionated Go testing harness for W3 Engineers projects,
built on top of [testify](https://github.com/stretchr/testify).

One import. One `New(t)` call. You get assertions, lifecycle hooks,
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

### Struct-based suite with lifecycle hooks

```go
import (
    testifywrapper "github.com/FarhanAsfar/testify-wrapper"
    "github.com/FarhanAsfar/testify-wrapper/suite"
)

type MyServiceSuite struct {
    suite.BaseSuite   // embed for no-op defaults — override only what you need
    db *sql.DB
}

func (s *MyServiceSuite) SetupSuite()    { s.db = connectTestDB() }
func (s *MyServiceSuite) TearDownTest()  { s.db.Exec("DELETE FROM orders") }
func (s *MyServiceSuite) Shutdown()      { s.db.Close() }

func (s *MyServiceSuite) TestCreateOrder() {
    // test logic here
}

// TestMyServiceSuite is the only function the Go test runner calls directly.
func TestMyServiceSuite(t *testing.T) {
    suite.Run(t, &MyServiceSuite{})
}
```

**Lifecycle order:**
```
SetupSuite
  SetupTest → TestXxx → TearDownTest   (once per Test* method)
TearDownSuite
Shutdown
```

### File-driven test

```go
func TestAdd_FileDriven(t *testing.T) {
    kit := testifywrapper.New(t)

    cases, err := kit.LoadJSON("testdata/cases.json")
    kit.Require().NoError(err)

    kit.RunCases(t, cases, func(t *testing.T, tc testifywrapper.TestCase) {
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

| Package                          | Purpose                                                     |
|----------------------------------|-------------------------------------------------------------|
| `testifyWrapper` (root)          | Entry point — `New(t)` returns an `Instance` with everything wired |
| `assert`                         | Non-fatal assertions — test keeps running after failure     |
| `require`                        | Fatal assertions — test stops immediately after failure     |
| `suite`                          | Struct-based suite runner with ordered lifecycle hooks      |
| `filehandler`                    | JSON/YAML fixture loader for data-driven tests              |
| `internal/hooks`                 | Internal cleanup registry (not for direct use)              |

---

## Fixture File Format

Both JSON and YAML are supported. Every fixture file must be a top-level
array where each entry has three fields:

| Field      | Type           | Description                                              |
|------------|----------------|----------------------------------------------------------|
| `name`     | string         | Becomes the subtest label in `go test` output            |
| `input`    | object         | Your input — unmarshal into your own struct in the callback |
| `expected` | object         | Your expected output — unmarshal into your own struct    |

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
`t.Cleanup` runs even when the test calls `t.Fatal` or panics. `defer` does
not guarantee this in all cases.

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
- Every new feature must have tests — no exceptions
- Update `CHANGELOG.md` under the appropriate version section

**Adding custom assertions:**
Domain-specific assertions (e.g. `AssertValidUUID`) belong in
`assert/assert.go` as methods on `Assertions`. Mirror the same method
in `require/require.go` for the fatal variant.