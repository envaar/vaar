<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# ending-blank-line

Flags .env files that do not end with exactly one final newline.

## What it catches

Missing trailing newlines and files that end with extra blank lines.

## Why it matters

Consistent file endings keep .env configurations clean.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking the file ending against the expected single blank line at the end.

## Fix

Leave one final newline and remove extra blank lines at the end.

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
