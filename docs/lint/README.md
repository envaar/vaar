<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Lint Guide

This page contains the lint-specific command reference for Vaar. Refer to [README](../README.md) for installation, verification and the shortest possible first-run steps.

## Description of the Command

`vaar lint` is the primary command for checking a repository for environment configuration issues. It reports findings with line numbers, supports deterministic fixes with `--fix`, machine-readable output with `--json` and explicit rule selection or exclusion with `--only` and `--skip` respectively.

To use lint, run the command:

- `vaar lint`

Lint also comes with additional flags such as:

- `vaar lint --fix`
- `vaar lint --json`
- `vaar lint --only=[rule-name]`
- `vaar lint --skip=[rule-name]`

## Lint Flags

### `--fix`

Applies only the safe formatting fixes that can be made deterministically.

### `--json`

Renders findings as a JSON output. To be used when a CI job, editor integration or wrapper script needs structured output instead of text.

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

- `0` means no lint findings were reported.
- `1` means lint findings were reported.
- `2` means the command failed before producing results.

The default output of `vaar lint` is plain text. Use `--json` when you want machine-readable findings for automation purposes.

## Rule Reference

The documentation for the specific rules followed by `vaar lint` are provided at [docs/lint/rules/README.md](./rules/README.md).
