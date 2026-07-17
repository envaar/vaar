<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Lint Guide

This page contains the lint-specific command reference for Vaar. Refer to [README](../README.md) for installation, verification and the shortest possible first-run steps.

## Description of the Command

`vaar lint` is the primary command for checking a repository for environment configuration issues. It reports findings with line numbers, supports deterministic fixes with `--fix`, machine-readable output with `--json`, JSON file export with `--output` or `-o`, explicit rule selection or exclusion with `--only` and `--skip` and explicit scope selection with either `--target` or `--target-dir`.

To use lint, run the command:

- `vaar lint`

Lint also comes with additional flags such as:

- `vaar lint --fix`
- `vaar lint --json`
- `vaar lint --json --output=lint-report.json`
- `vaar lint --json -o lint-report.json`
- `vaar lint --only=[rule-name]`
- `vaar lint --skip=[rule-name]`
- `vaar lint --target=.env.staging`
- `vaar lint --target-dir=src`
- `vaar lint --list-rules`

## Lint Flags

### `--fix`

Applies only the safe formatting fixes that can be made deterministically. Vaar reports the findings from the original file and marks findings that disappeared after the fix with `[fixed]`. It then reports any findings that remain in the post-fix file. The exit code is based on the remaining findings, so a file with only repaired findings exits successfully.

### `--json`

Renders findings as a JSON output. To be used when a CI job, editor integration or wrapper script needs structured output instead of text.

### `--output`, `-o`

Writes the JSON report to a destination file path instead of `stdout`. This flag requires `--json`.

- Relative and absolute destination paths are both supported.
- The parent directory must already exist; `vaar lint` does not create missing directories.
- Existing destination files are replaced.
- The filename extension does not control the reporter; `--output=report.txt` still writes JSON.
- Exit codes continue to report the lint result, not just whether the file write succeeded.
- If the resolved output path matches a lint input file, `vaar lint` fails before writing output.

Examples:

```bash
vaar lint --json --output=lint-report.json
vaar lint --json -o lint-report.json
vaar lint --json --output=reports/lint-report.json
vaar lint \
  --target=.env.staging \
  --json \
  --output=reports/staging-lint.json
```

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

### `--list-rules`

Prints every registered rule in alphabetical order with its canonical name and a concise description, then exits successfully. This is an informational mode: it does not discover dotenv files, run rules or require a repository, so it works from any directory.

The output uses NAME and DESCRIPTION columns. A FIXABLE column is not included because the rule registry does not currently expose fixability metadata.

`--list-rules` cannot be combined with execution or output flags: `--only`, `--skip`, `--fix`, `--target`, `--target-dir`, `--output` or `--json`; doing so returns a usage error naming the conflicting flag.

Example:

```bash
vaar lint --list-rules
```

## Output and Exit Codes

`vaar lint` uses simple exit codes that are useful for scripting:

- `0` means no lint findings were reported (if `--fix` made any fixes, they are reported)
- `1` means lint findings were reported.
- `2` means the command failed before producing results.

The default output of `vaar lint` is plain text. Repaired findings are prefixed with `[fixed]`. Use `--json` when you want machine-readable findings for automation purposes; repaired findings include `"fixed": true`.

When `--output` is not used, JSON is written to `stdout`. When `--output` or `-o` is used, the JSON report is written to the destination file instead and is not printed to `stdout`.

Producing a report file and passing lint are separate outcomes:

- If lint finds no issues, `vaar lint --json --output=report.json` writes the JSON report and exits with code `0`.
- If lint finds issues, `vaar lint --json --output=report.json` still writes the JSON report and exits with code `1`.
- If `--output` is used without `--json`, if the destination cannot be written, if the parent directory does not exist or if the output path resolves to a lint input file, `vaar lint` fails before producing the report and exits with code `2`.

## Rule Reference

The documentation for the specific rules followed by `vaar lint` are provided at [docs/lint/rules/README.md](./rules/README.md).
