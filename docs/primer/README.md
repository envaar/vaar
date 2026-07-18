<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Vaar Primer

This page is a short introduction to Vaar for new developers. It explains what the project does, how a lint execution moves through the codebase and how to run the tool locally.

This primer complements the reference docs: [System Overview](../system-overview.md)
for package boundaries, [Usage](../usage.md) for the command map and
[Lint Guide](../lint/README.md) for the full flag and exit-code reference.

## Foundations

### What Vaar Is

Vaar is a Go command-line tool for checking environment configuration. It discovers dotenv files in a repository, runs the built-in rule set, reports findings with file and line numbers and applies safe formatting fixes with `--fix`.

### The Problem It Solves

`.env` files are rarely reviewed with the same rigor as code. Because they are sensitive, they are difficult to share and diff. So duplicate keys, wrong delimiters, stray whitespaces, non-portable key names and other such issues slip through unnoticed. Vaar makes that drift visible and, where a fix is unambiguous, repairs it.

### What Vaar Checks 

By default Vaar discovers `.env` and any filename beginning with `.env.`, such
as `.env.example`, `.env.local` or `.env.preview-local`. It does not treat
`.env.`, `.environment`, `.envrc` or `my.env` as dotenv files.

Discovery is recursive and stays deterministic. The repository walk skips generated, vendored, fixture and VCS directories, as implemented in `internal/fs/`.

### Design Principles

These principles detailed in the [System Overview](../system-overview.md) and are visible in the package layout:

- Deterministic first. Every rule is currently under `internal/lint/rules/deterministic/` and reports an exact finding from file content.
- Stable ordering and output. Findings are sorted by file, line, severity, rule and message so scripts and CI can rely on the result.
- Preserve original file state where possible. Parsing keeps original bytes, line numbers, BOM state and line-ending information available to later stages.
- Keep the command layer thin. `internal/cli/` handles command wiring, flags and exit codes while the main logic stays in focused packages.

## Components

A lint run passes through a small set of packages. Each package has one main
responsibility.

| Stage | Package | Responsibility |
| ----- | ------- | -------------- |
| Entrypoint | `cmd/vaar/` | Starts the binary and hands control to `internal/cli`. |
| CLI layer | `internal/cli/` | Cobra command wiring, flag translation, exit codes and user-facing errors. |
| Discovery / walk | `internal/fs/` | Walks the tree, matches known dotenv filenames and skips ignored directories such as `.git`, `build`, `dist`, `node_modules`, `testdata` and `vendor` (`Discover`). |
| Envfile parser | `internal/envfile/` | Parses bytes into a line-aware model and provides `Normalize`/`Write` for safe rewrites. |
| Rule engine | `internal/lint/` | Selects rules with `--only`/`--skip`, runs them, sorts findings and drives the fix pass (`Runner`). |
| Rule registry | `internal/lint/rules/` | `All()` returns the built-in rule set in a stable order over the category packages. |
| Report / output | `internal/report/` | Renders findings as plain text or JSON. |

### How A Lint Run Works

The `Runner` in `internal/lint/runner.go` ties these packages together:

1. Select the active rules after applying `--only` and `--skip`.
2. Resolve the lint scope: the whole repository, one `--target` file or one `--target-dir` tree.
3. Discover dotenv files and load them into in-memory snapshots.
4. Run the selected rules against those snapshots.
5. If `--fix` is enabled, apply safe fixes, re-parse the changed files, re-run the rules and mark findings that disappeared as `[fixed]`.
6. Sort the findings into a stable order.
7. Render text or JSON output.
8. Exit based on the findings that remain after fixing.

Each finding carries a severity (`warn` or `error`), a rule name, the file, a line number and a message. The built-in rules are catalogued in [docs/lint/rules/README.md](../lint/rules/README.md).

## Getting Started

### Build And Verify

For install-from-release and `go install` paths, see the [README Installation section](../../README.md#installation). To build from a clone:

```bash
make build
./bin/vaar --version
```

```text
vaar version dev
```

A source build reports the version as `dev`. Release binaries report the release tag.

The examples below use `./bin/vaar` from a local clone. If `vaar` is already on your `PATH`, use the same commands without the `./bin/` prefix.

### First Lint Run

From the Vaar repository, run the linter against the [broken example](../../examples/broken/README.md):

```bash
./bin/vaar lint --target=examples/broken/.env.example
```

It reports:

```text
warn space-character examples/broken/.env.example:2 line has spaces around the key, delimiter or value
warn trailing-whitespace examples/broken/.env.example:2 line has trailing whitespace
error duplicate-key examples/broken/.env.example:4 APP_ENV is defined more than once
error incorrect-delimiter examples/broken/.env.example:5 DATABASE_URL uses ':' instead of '='
error invalid-key-name examples/broken/.env.example:6 api-key is not a portable env key name
warn ending-blank-line examples/broken/.env.example:8 file must end with exactly one final newline
warn extra-blank-line examples/broken/.env.example:8 repeated blank line
```

### Reading The Output

Every line follows `<severity> <rule> <file>:<line> <message>`; during `--fix` runs, repaired findings additionally carry a `[fixed]` prefix (see below). A clean run prints nothing. Exit codes are script-friendly:

- `0` - no findings remain (fixes made during `--fix` are still reported).
- `1` - findings remain.
- `2` - the command failed before producing results (for example, an unknown rule name).

### JSON Output

Use `--json` when you need structured findings. Add `--output`/`-o` to write the JSON to a file instead of stdout (it requires `--json`):

```bash
./bin/vaar lint --target=examples/broken/.env.example --json
./bin/vaar lint --target=examples/broken/.env.example --json --output=findings.json
```

```json
{
  "findings": [
    {
      "rule": "duplicate-key",
      "severity": "error",
      "file": "examples/broken/.env.example",
      "line": 4,
      "message": "APP_ENV is defined more than once"
    }
  ]
}
```

### Applying Safe Fixes

`--fix` normalizes any findings that can be repaired safely, then re-checks the file. Repaired findings are prefixed with `[fixed]`; anything that still needs human intervention (such as a duplicate key or a wrong delimiter) remains:

```bash
./bin/vaar lint --target=examples/broken/.env.example --fix
```

```text
warn space-character examples/broken/.env.example:2 line has spaces around the key, delimiter or value
[fixed] warn trailing-whitespace examples/broken/.env.example:2 line has trailing whitespace
error duplicate-key examples/broken/.env.example:4 APP_ENV is defined more than once
error incorrect-delimiter examples/broken/.env.example:5 DATABASE_URL uses ':' instead of '='
error invalid-key-name examples/broken/.env.example:6 api-key is not a portable env key name
[fixed] warn ending-blank-line examples/broken/.env.example:8 file must end with exactly one final newline
[fixed] warn extra-blank-line examples/broken/.env.example:8 repeated blank line
```

### Selecting Rules

`--only` runs just the chosen rules; `--skip` removes rules specific rules. Both flags are repeatable and an unknown rule name fails the run before linting starts:

```bash
./bin/vaar lint --target=examples/broken/.env.example --only=duplicate-key
./bin/vaar lint --target=examples/broken/.env.example --skip=trailing-whitespace --skip=extra-blank-line
```

```text
error duplicate-key examples/broken/.env.example:4 APP_ENV is defined more than once
```

### Choosing Search Scope

By default, Vaar walks the whole repository. Narrow the scope with `--target` for one file or `--target-dir` for one tree; the two are mutually exclusive:

```bash
vaar lint --target=.env.staging
vaar lint --target-dir=src
```

## Where To Go Next

- Every flag, output detail and exit code: [Lint Guide](../lint/README.md).
- The full command map: [Usage](../usage.md).
- What each rule catches, with good and bad examples: [Rule Reference](../lint/rules/README.md).
- Package boundaries and where new code belongs: [System Overview](../system-overview.md).
- Contributing a change or a new rule: [CONTRIBUTING.md](../../CONTRIBUTING.md).
