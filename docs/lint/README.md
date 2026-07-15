<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Lint Guide

This page contains the lint-specific command reference for Vaar. Refer to [README](../README.md) for installation, verification and the shortest possible first-run steps.

## Description of the Command

`vaar lint` is the primary command for checking a repository for environment configuration issues. It reports findings with line numbers, supports deterministic fixes with `--fix`, machine-readable output with `--json`, explicit rule selection or exclusion with `--only` and `--skip` and explicit scope selection with either `--target` or `--target-dir`.

To use lint, run the command:

- `vaar lint`

Lint also comes with additional flags such as:

- `vaar lint --fix`
- `vaar lint --json`
- `vaar lint --only=[rule-name]`
- `vaar lint --skip=[rule-name]`
- `vaar lint --target=.env.staging`
- `vaar lint --target-dir=src`

## Lint Flags

### `--fix`

Applies only the safe formatting fixes that can be made deterministically. Vaar reports the findings from the original file and marks findings that disappeared after the fix with `[fixed]`. It then reports any findings that remain in the post-fix file. The exit code is based on the remaining findings, so a file with only repaired findings exits successfully.

### `--json`

Renders findings as a JSON output. To be used when a CI job, editor integration or wrapper script needs structured output instead of text.

### `--target`

Lints only the specified file path. The path can be relative or absolute, and the file does not need to match the default dotenv filename list. 

`--target` and `--target-dir` are mutually exclusive. If both `--target` and `--target-dir` are supplied, `vaar lint` returns an error.

### `--target-dir`

Recursively discovers dotenv files under the specified directory while keeping the normal repository ignore behavior.

`--target` and `--target-dir` are mutually exclusive. If both `--target` and `--target-dir` are supplied, `vaar lint` returns an error.

### `--only`

Useful for selecting a specific rule or a specific list of rules to run.

- The flag can be repeated to specify each rule to be checked.
- Unknown rule names are rejected before linting starts.
- Listing the same rule more than once does not run it multiple times.

Examples:

```bash
vaar lint --only=duplicate-key
vaar lint --only=duplicate-key --only=invalid-key-name
```

### `--skip`

Useful for selecting a specific rule or a specific list of rules to be skipped.

- The flag can be repeated to specify each rule to be skipped.
- Unknown rule names are rejected before linting starts.

Examples:

```bash
vaar lint --skip=trailing-whitespace
vaar lint --skip=trailing-whitespace --skip=extra-blank-line
```

## Output and Exit Codes

`vaar lint` uses simple exit codes that are useful for scripting:

- `0` means no lint findings were reported (if `--fix` made any fixes, they are reported)
- `1` means lint findings were reported.
- `2` means the command failed before producing results.

The default output of `vaar lint` is plain text. Repaired findings are prefixed with `[fixed]`. Use `--json` when you want machine-readable findings for automation purposes; repaired findings include `"fixed": true`.

## Rule Reference

The documentation for the specific rules followed by `vaar lint` are provided at [docs/lint/rules/README.md](./rules/README.md).
