<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Vaar Primer

A fast, hands-on tour of Vaar for new developers. Read this end to end and you
will know what Vaar is, how a lint run flows through the codebase and how to run
it yourself. It should take about ten minutes.

This primer complements the reference docs: [System Overview](../system-overview.md)
for package boundaries, [Usage](../usage.md) for the command map and
[Lint Guide](../lint/README.md) for the full flag and exit-code reference.

## Foundations

### What Vaar Is

Vaar is a repo-aware linter for environment variables, shipped as a single Go
binary (`cmd/vaar`). The current release focuses on **deterministic dotenv
hygiene**: it discovers `.env` files in a repository, checks each one against a
built-in rule set, reports findings with file and line numbers and can apply
safe formatting fixes.

### The Problem It Solves

`.env` files are rarely reviewed with the same rigor as code. Because they are
sensitive, they are awkward to share and diff, so duplicate keys, wrong
delimiters, stray whitespace and non-portable key names slip through unnoticed.
Vaar makes that drift visible and, where a fix is unambiguous, repairs it.

### Design Principles

These principles come straight from the [System Overview](../system-overview.md)
and are visible in the package layout:

- **Deterministic first.** Every shipped rule reports an exact finding proven
  from the file's own bytes. Rules live under `internal/lint/rules/deterministic/`.
- **Stable ordering and output.** Findings are sorted by file, line, severity,
  rule and message so scripts and CI can rely on the result (`sortFindings` in
  `internal/lint/runner.go`).
- **Preserve original state.** The parser keeps original bytes, line numbers,
  BOM state and line-ending information; a file is only rewritten when a safe
  fix intentionally changes it (`internal/envfile/`, `internal/lint/fix.go`).
- **Thin command layer.** `internal/cli/` only wires flags, exit codes and
  output; the real logic lives in focused packages.
- **Local and self-contained.** Vaar reads files from the working tree and
  nothing else — no network calls, no API keys, no daemon.

## Components

A lint run flows through a small set of packages. Each one owns a single
responsibility, which keeps them easy to test and extend.

| Stage | Package | Responsibility |
| ----- | ------- | -------------- |
| Entrypoint | `cmd/vaar/` | Starts the binary and hands control to `internal/cli`. |
| CLI layer | `internal/cli/` | Cobra command wiring, flag translation, exit codes and user-facing errors. |
| Discovery / walk | `internal/fs/` | Walks the tree, matches known dotenv filenames and skips build, VCS and vendored directories (`Discover`). |
| Envfile parser | `internal/envfile/` | Parses bytes into a line-aware model and provides `Normalize`/`Write` for safe rewrites. |
| Rule engine | `internal/lint/` | Selects rules with `--only`/`--skip`, runs them, sorts findings and drives the fix pass (`Runner`). |
| Rule registry | `internal/lint/rules/` | `All()` returns the built-in rule set in a stable order over the category packages. |
| Report / output | `internal/report/` | Renders findings as plain text or JSON. |

### How A Lint Run Works

The `Runner` in `internal/lint/runner.go` ties these together:

1. Select the active rules (apply `--only`, then `--skip`).
2. Resolve the scope: the whole repo, one `--target` file or one `--target-dir`
   tree.
3. Discover dotenv files and parse each into an in-memory snapshot.
4. Run every selected rule against the snapshots to produce findings.
5. If `--fix` is set, apply safe fixes, re-parse, re-run the rules and mark
   findings that disappeared as `[fixed]`.
6. Sort the findings into a stable order.
7. Render text or JSON and choose the exit code from the remaining findings.

Each finding carries a severity (`warn` or `error`), a rule ID, the file, a line
number and a message. The built-in rules are catalogued in
[docs/lint/rules/README.md](../lint/rules/README.md).

## Getting Started

### Build And Verify

For install-from-release and `go install` paths, see the
[README Installation section](../../README.md#installation). To build from a
clone:

```bash
make build
./bin/vaar --version
```

```text
vaar version dev
```

A source build reports the version as `dev`; release binaries report their tag.

### Your First Lint Run

Run Vaar from the repository you want to check. Against the
[broken example](../../examples/broken/README.md) it reports:

```text
warn space-character .env:2 line has spaces around the key, delimiter or value
warn trailing-whitespace .env:2 line has trailing whitespace
error duplicate-key .env:4 APP_ENV is defined more than once
error incorrect-delimiter .env:5 DATABASE_URL uses ':' instead of '='
error invalid-key-name .env:6 api-key is not a portable env key name
warn ending-blank-line .env:8 file must end with exactly one final newline
warn extra-blank-line .env:8 repeated blank line
```

### Reading The Output

Every line follows `<severity> <rule> <file>:<line> <message>`. A clean run
prints nothing. Exit codes are script-friendly:

- `0` — no findings remain (fixes made during `--fix` are still reported).
- `1` — findings remain.
- `2` — the command failed before producing results (for example, an unknown
  rule name).

### JSON Output

Use `--json` when a CI job or wrapper needs structured findings. Add
`--output`/`-o` to write the JSON to a file instead of stdout (it requires
`--json`):

```bash
vaar lint --json
vaar lint --json --output=findings.json
```

```json
{
  "findings": [
    {
      "rule": "duplicate-key",
      "severity": "error",
      "file": ".env",
      "line": 4,
      "message": "APP_ENV is defined more than once"
    }
  ]
}
```

### Applying Safe Fixes

`--fix` normalizes what can be repaired deterministically, then re-checks the
file. Repaired findings are prefixed with `[fixed]`; anything that still needs a
human decision (such as a duplicate key or a wrong delimiter) remains:

```bash
vaar lint --fix
```

```text
warn space-character .env:2 line has spaces around the key, delimiter or value
[fixed] warn trailing-whitespace .env:2 line has trailing whitespace
error duplicate-key .env:4 APP_ENV is defined more than once
error incorrect-delimiter .env:5 DATABASE_URL uses ':' instead of '='
error invalid-key-name .env:6 api-key is not a portable env key name
[fixed] warn ending-blank-line .env:8 file must end with exactly one final newline
[fixed] warn extra-blank-line .env:8 repeated blank line
```

### Selecting Rules

`--only` runs just the named rules; `--skip` removes rules after selection. Both
flags are repeatable, and an unknown rule name fails the run before linting
starts:

```bash
vaar lint --only=duplicate-key
vaar lint --skip=trailing-whitespace --skip=extra-blank-line
```

```text
error duplicate-key .env:4 APP_ENV is defined more than once
```

### Choosing A Scope

By default Vaar walks the whole repository. Narrow the scope with `--target` for
one file or `--target-dir` for one tree; the two are mutually exclusive:

```bash
vaar lint --target=.env.staging
vaar lint --target-dir=src
```

## Where To Go Next

- **Every flag, output detail and exit code:** [Lint Guide](../lint/README.md).
- **The full command map:** [Usage](../usage.md).
- **What each rule catches, with good and bad examples:**
  [Rule Reference](../lint/rules/README.md).
- **Package boundaries and where new code belongs:**
  [System Overview](../system-overview.md).
- **Contributing a change or a new rule:**
  [CONTRIBUTING.md](../../CONTRIBUTING.md).
