<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Usage Guide

This page is the top-level guide to using Vaar from the command line.
For installation, verification and the fastest first run, start with [README](../README.md).

## Command Map

| Command                                                                          | What it does                            | Read more                                          |
| -------------------------------------------------------------------------------- | --------------------------------------- | -------------------------------------------------- |
| `vaar lint`                                                                      | Runs the repository lint checks.        | [Lint guide](./lint/README.md)                     |
| `vaar lint --fix` / `vaar lint --json` / `vaar lint --only` / `vaar lint --skip` | Adjusts lint output and rule selection. | [Lint guide](./lint/README.md)                     |
| `vaar --help` / `vaar help`                                                      | Shows the root help screen.             | [Help guide](./help/README.md)                     |
| `vaar lint --help` / `vaar help lint`                                            | Shows the lint help screen.             | [Lint help guide](./help/help-lint.md)             |
| `vaar completion <shell>`                                                        | Prints a shell completion script.       | [Completion guide](./completion/README.md)         |
| `vaar completion --help` / `vaar help completion`                                | Shows the completion help screen.       | [Completion help guide](./help/help-completion.md) |
| `vaar --version`                                                                 | Prints the installed version.           | [README](../README.md)                             |

## Where To Go Next

- Need the lint command reference, flags, exit codes or rule catalog? Read [docs/lint/README.md](./lint/README.md).
- Need rule-by-rule behavior and examples? Read [docs/lint/rules/README.md](./lint/rules/README.md).
- Need the root help screen or help-screen behavior? Read [docs/help/README.md](./help/README.md).
- Need shell setup instructions for completions? Read [docs/completion/README.md](./completion/README.md).

## Common Paths

1. Install Vaar and verify the binary from [README](../README.md).
2. Run `vaar lint` in the repository you want to check.
3. Open [docs/lint/README.md](./lint/README.md) when you need to use/exclude specific rules, use `--json` or understand exit codes.
