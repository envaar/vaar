<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# substitution-key

Flags malformed `$KEY` and `${KEY}` substitution syntax inside assignment values.

## Problem

Dotenv parsers commonly recognize `$KEY` and `${KEY}` inside values. A missing or extra brace can make the value behave like a substitution or literal text.

## What it catches

Malformed substitutions in unquoted and double-quoted assignment values.

The rule reports:

- `${KEY` missing its closing `}`
- unbraced substitutions like `$KEY}` with an unmatched closing `}`
- `${}` with an empty key
- `${key-name}` when the key is not a portable environment variable name

> [!NOTE]
> Single-quoted values are literal text and are not scanned. Inline comments
> recognized by the parser are also ignored.

## Why it matters

Malformed substitutions can behave differently across dotenv parsers, shells and
deployment tools.

## Evidence level

- `Deterministic`: Vaar can prove this from the parsed file contents alone by detecting malformed `$KEY` and `${KEY}` syntax in assignment values, without expanding the value or checking whether the referenced key exists.

## Fix

Use one of the supported substitution forms: `$KEY` or `${KEY}`.

Automatic fixing is not currently provided because adding or removing braces can
change the value's meaning.

## Bad example

```dotenv
ABC=${BAR
FOO="$BAR}"
EMPTY_REFERENCE=${}
```

## Good example

```dotenv
ABC=${BAR}
FOO="$BAR"
BRACED=${BAR}
UNBRACED=$BAR
QUOTED_BRACED="${BAR}"
QUOTED_UNBRACED="$BAR"
MIXED="prefix-${BAR}-$BAZ-suffix"
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=substitution-key
```

Bad Example Produces:

```text
error substitution-key .env:1 substitution "${BAR" is missing a closing "}"
error substitution-key .env:2 substitution "$BAR}" contains an unmatched closing "}"
error substitution-key .env:3 substitution "${}" is empty
```
