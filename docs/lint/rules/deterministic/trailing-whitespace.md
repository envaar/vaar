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
KEY=value\s
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=trailing-whitespace
```

Bad Example Produces:

```text
warn trailing-whitespace .env:1 line has trailing whitespace
```

> [!NOTE]
> `\s` above stands for an actual trailing space, which Markdown cannot show.
> A file containing the literal two characters `\s` has no trailing
> whitespace and does not produce this finding.
