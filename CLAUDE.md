# mk — Go port of Plan 9 mk

A build tool reimplementing Plan 9's `mk` in Go, with improvements (parallel execution, pcre-like regex rules, flexible indentation, alternative recipe shells).

## Quick Reference

```bash
go build          # build ./mk binary
go test           # run all tests
go test -run TestScript/basic_rule  # run a single script test
```

## Architecture

Single `package main`, procedural style. Pipeline: **Lexer → Parser → Graph → Executor**.

| File | Purpose |
|------|---------|
| `mk.go` | Entry point (`main`), CLI flags, parallel build orchestration |
| `lex.go` | Lexer — channel-based token stream, state machine |
| `parse.go` | Parser — recursive descent, executes assignments on-the-fly |
| `expand.go` | Variable/backtick expansion, pattern substitution |
| `graph.go` | Dependency graph construction, cycle detection |
| `recipe.go` | Recipe execution, shell subprocess management |
| `rules.go` | Rule structs, attributes (D,E,N,n,Q,R,U,V,X,S,P) |

## Test Pattern

Integration tests use `rsc.io/script/scripttest` with txtar files in `testdata/*.txt`. Each file is a self-contained test combining script commands, mkfile content, and expected output.

Tests use `TestMain` with `TEST_MAIN=mk` env var to re-exec the test binary as `mk` itself. The `mk` command is registered in the script engine via `script.Program`.

To add a test: create `testdata/descriptive_name.txt`:
```txtar
# Description of what this tests
mk -n -f mkfile
cmp stdout expected

-- mkfile --
target:
	recipe
-- expected --
target: recipe
```

Run a single script test: `go test -run TestScript/basic_rule`

A few tests remain as Go tests in `mk_test.go`: `TestRecipesHaveEnv` (programmatic env inspection) and `TestInteractiveMode*` (stdin piping).

## Pre-commit Checklist

Before committing, all of the following must pass:

```bash
go test ./...              # all tests must pass
go vet ./...               # must be clean
go tool staticcheck ./...  # must be clean (pinned in go.mod)
go tool govulncheck ./...  # no known vulnerabilities (pinned in go.mod)
```

Address any gopls diagnostics (type errors, unused imports, etc.) visible in changed files.

All changed/added code must be covered by tests. Check uncovered lines:
```bash
go test -coverprofile=cover.out ./...; awk -f cover-uncovered.awk cover.out
```
Output is one line per file with uncovered line ranges (e.g. `expand.go: 54-58,90-92`).

## Code Review

Before finalizing a change, review the diff with Sonnet. Focus the review on things mechanical checks can't catch: correctness, test quality, design issues, flaky tests, and subtle bugs. Don't duplicate work already covered by the pre-commit checklist (test runs, vet, staticcheck, coverage).

## Conventions

- No external runtime dependencies — only stdlib at runtime.
- Errors use `mkError()` (fprintf to stderr + os.Exit(1)), not panic.
- Global state for color, shell, rebuild flags.
- `\x01` separates array elements in environment variables (Plan 9 rc convention).
