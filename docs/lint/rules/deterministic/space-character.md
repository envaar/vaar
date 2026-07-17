<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# space-character

Flags spaces around the key, delimiter or value.

## What it catches

Whitespace that weakens dotenv portability or makes the file inconsistent.

## Why it matters

Spacing differences can hide subtle parsing behavior.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by matching whitespace around the key, delimiter or value.

## Fix

Normalize the spacing around the assignment.

## Bad example

```dotenv
KEY = value
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=space-character
```

Bad Example Produces:

```text
warn space-character .env:1 line has spaces around the key, delimiter or value
```
