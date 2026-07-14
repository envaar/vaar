<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# extra-blank-line

Flags repeated or unecessary extra blank lines between environment variables.

## What it catches

Multiple blank lines in a row between variables.

## Why it matters

It makes small dotenv files harder to scan and review.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by detecting consecutive blank lines between variables or at the end.

## Fix

Combine repeated blank lines to a single blank line.

> [!NOTE]
> This is also fixed automatically by running `vaar lint --fix`.

## Bad example

```dotenv
KEY=value


OTHER=value
```

## Good example

```dotenv
KEY=value

OTHER=value
```
