<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# trailing-whitespace

Flags trailing spaces or tabs at the end of a line.

## What it catches

Invisible whitespace after the line content.

## Why it matters

It creates noisy diffs and can hide formatting mistakes.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking for spaces or tabs at the end of a line.

## Fix

Trim the trailing whitespace.

> [!NOTE]
> This is also fixed automatically by running `vaar lint --fix`.

## Bad example

```dotenv
KEY=value
```

## Good example

```dotenv
KEY=value
```
