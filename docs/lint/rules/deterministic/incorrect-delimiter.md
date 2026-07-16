<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# incorrect-delimiter

Flags dotenv lines that use `:` or other such non-standard delimiters instead of `=`.

## What it catches

Key-value lines written with the wrong separator.

## Why it matters

Different parsers may interpret the line differently or reject it entirely. Useful for uniformity.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by detecting non-standard delimiters where `=` is expected.

## Fix

Replace the non-standard delimiter with `=`.

## Bad example

```dotenv
KEY:value
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=incorrect-delimiter
```

Bad Example Produces:

```text
error incorrect-delimiter .env:1 KEY uses ':' instead of '='
```
