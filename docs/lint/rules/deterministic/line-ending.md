<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# line-ending

Flags mixed CRLF and LF line endings in the same file.

## What it catches

Files that mix Windows and Unix line endings.

## Why it matters

Mixed line endings create noisy diffs and inconsistent behavior across tools.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking the line-ending bytes in the file.

## Fix

Normalize the file to one line-ending style.

> [!NOTE]
> This is also fixed automatically by running `vaar lint --fix`.

## Bad example

```text
KEY=value\r\nOTHER=value\n
```

## Good example

```text
KEY=value\nOTHER=value\n
```
