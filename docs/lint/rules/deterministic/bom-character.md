<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# bom-character

Flags files that start with a UTF-8 BOM.

## What it catches

An invisible byte-order mark at the start of the file.

## Why it matters

Some tools handle BOMs differently, which can create confusing parse issues.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking whether the file starts with a UTF-8 BOM.

## Fix

Remove the BOM.

> [!NOTE]
> This is also fixed automatically by running `vaar lint --fix`.

## Bad example

```text
\ufeffKEY=value
```

## Good example

```text
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=bom-character
```

Bad Example Produces:

```text
warn bom-character .env:1 file starts with a UTF-8 BOM
```

> [!NOTE]
> `\ufeff` above stands for the actual BOM bytes (`EF BB BF`) at the start of the
> file. A file containing the literal six characters `\ufeff` is not a BOM and
> does not produce this finding.
