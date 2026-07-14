<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Completion Guide

This page contains the completion-specific command reference for Vaar. Refer to [README](../../README.md) for installation, verification and the shortest possible first-run steps.

## Description of the Command

`vaar completion` prints a Cobra-generated shell completion script for the shell you choose. The generated script stays in sync with the latest command tree and with the lint rule names for easier autocompletion in cases like `vaar lint --only` and `vaar lint --skip`.

The command supports the following shells:

- `bash`
- `zsh`
- `fish`
- `powershell`

To generate a script, run:

- `vaar completion bash`
- `vaar completion zsh`
- `vaar completion fish`
- `vaar completion powershell`

The script is written to standard output, so you can source it in the current shell session or redirect it into the shell's completion directory.

## Shell Setup

### Bash

`vaar completion bash` depends on the `bash-completion` package.

To load completions in the current shell session:

```bash
source <(vaar completion bash)
```

To install them permanently:

```bash
vaar completion bash > /etc/bash_completion.d/vaar
```

On macOS:

```bash
vaar completion bash > $(brew --prefix)/etc/bash_completion.d/vaar
```

You will need to start a new shell after installing the script.

### Zsh

If shell completion is not already enabled, run once:

```zsh
autoload -U compinit; compinit
```

To load completions in the current shell session:

```zsh
source <(vaar completion zsh)
```

To install them permanently on Linux:

```zsh
vaar completion zsh > "${fpath[1]}/_vaar"
```

On macOS:

```zsh
vaar completion zsh > $(brew --prefix)/share/zsh/site-functions/_vaar
```

You will need to start a new shell after installing the script.

### Fish

To load completions in the current shell session:

```fish
vaar completion fish | source
```

To install them permanently:

```fish
vaar completion fish > ~/.config/fish/completions/vaar.fish
```

You will need to start a new shell after installing the script.

### PowerShell

To load completions in the current shell session:

```powershell
vaar completion powershell | Out-String | Invoke-Expression
```

To install them permanently, add the output of that command to your PowerShell profile.

You will need to start a new shell after installing the script.

## Completion Flags

### `--no-descriptions`

Disables completion descriptions in the generated script.

Example:

```bash
vaar completion bash --no-descriptions
```

## Lint Rule Completions

`vaar lint` also registers completions for:

- `--only`
- `--skip`

Those flags suggest the current built-in rule names and disable file completion for the flag values.

Examples:

```bash
vaar lint --only=<TAB>
vaar lint --skip=<TAB>
```

The suggestions come from the latest rule set that is available with Vaar, so they stay aligned with the most current version of the codebase.
They are returned in declaration order, which keeps the completion list stable as the rule set grows.

## Running From Source

If you are trying completions before installing a release binary, the same commands work with:

```bash
go run ./cmd/vaar completion <shell>
```

## Output and Exit Behavior

`vaar completion` writes the completion script to standard output and exits successfully for supported shells. The shell-specific subcommands show Cobra's own usage text when you pass `--help`.
