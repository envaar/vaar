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

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=line-ending
```

Bad Example Produces:

```text
warn line-ending .env:1 file uses mixed CRLF and LF line endings
```

> [!NOTE]
> `\r\n` and `\n` above stand for the actual line-ending bytes, so the bad
> example is a two-line file whose first line ends with CRLF and whose second
> line ends with LF.
