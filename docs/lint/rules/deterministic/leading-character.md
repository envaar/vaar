<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# leading-character

Flags leading whitespace before a dotenv key.

## What it catches

Indented assignment lines that are likely accidental.

## Why it matters

Leading whitespace can hide formatting problems and make files harder to review.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking for leading whitespace before a dotenv key.

## Fix

Remove the leading indentation.

## Bad example

```dotenv
 KEY=value
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=leading-character
```

Bad Example Produces:

```text
warn leading-character .env:1 line starts with leading whitespace
```
