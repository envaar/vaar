<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# System Overview

Vaar keeps the command layer thin and pushes the real logic into a few focused packages. That makes the project easier to test, easier to extend and easier to keep stable as it grows. This page explains the current layout, the lint pipeline and the right place for new work. For command usage, see [Usage](./usage.md).

## Design Principles

- Keep the command layer thin and declarative in nature.
- Keep parsing, writing and line-aware file modeling in `internal/envfile/`.
- Make repository discovery explicit and deterministic.
- Prefer stable ordering and stable output so that scripts can rely on results.
- Preserve original environment file state unless a safe fix intentionally changes them.
- Treat documentation as part of the public contract that has to be maintained.

## Package Map

| Package                              | Responsibility                     | Notes                                                                 |
| ------------------------------------ | ---------------------------------- | --------------------------------------------------------------------- |
| `cmd/vaar/`                          | Program entrypoint                 | Starts the binary and hands control to `internal/cli`.                |
| `internal/cli/`                      | Command wiring and exit codes      | Handles Cobra setup, flag translation and user-facing error handling. |
| `internal/fs/`                       | Repository discovery               | Finds dotenv files and skips generated, vendored and fixture paths.   |
| `internal/envfile/`                  | Parsing and file modeling          | Preserves line-level detail, original bytes and newline state.        |
| `internal/lint/`                     | Lint orchestration                 | Selects rules, runs them, sorts findings and applies safe fixes.      |
| `internal/lint/rules/`               | Rule catalog                       | Exposes `All()` and compatibility helpers for the rule packages.      |
| `internal/lint/rules/deterministic/` | Deterministic rule implementations | One file per rule along with focused tests and docs.                  |
| `internal/report/`                   | Output rendering                   | Produces text and JSON output for humans and automation.              |

## How a Lint Run Works

1. Resolve the repository root.
2. Discover dotenv files in the tree.
3. Create an in-memory snapshot of each discovered `.env` file.
4. Select the rules to apply.
5. Parse each file and run the rules against the original snapshot.
   1. If automatic safe fixing is enabled with `--fix`:
      1. Apply the safe fixes.
      2. Re-parse the files and re-run the rules after fixing, marking disappeared findings as fixed while retaining findings that remain.
6. Sort findings into a stable order.
7. Render text or JSON output.
8. Exit based on the post-fix findings.

The parser keeps the original bytes, line numbers, BOM state and line-ending information available to later stages. This ensures safe rewrites.

## Where to Add New Functionality/Code

- New command or flag behavior: `internal/cli/`
- Repository walking or ignore logic: `internal/fs/`
- Parsing, formatting or write-back semantics: `internal/envfile/`
- Rule selection or rule orchestration: `internal/lint/`
- A new deterministic lint rule: `internal/lint/rules/deterministic/`
- Reporter changes: `internal/report/`
- User-facing documentation: `README.md`, `docs/usage.md`, `docs/lint/README.md` or the relevant rule page

If a change affects user-facing behaviour, add or update tests and document the behaviour appropriately.

> [!NOTE]
> Please note that the CLI command set, documented flags, exit codes, output formats and shipped rule behaviour should remain stable. While internal package names may evolve or change, their behaviour should stay stable and documented once released. Breaking changes should ensure backwards compatibility wherever possible or proper migration paths wherever not.
