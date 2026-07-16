<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# constant-case

Warns about portable keys that are not CONSTANT_CASE.

## What it catches

Keys that match the portable env-key pattern but contain lowercase ASCII
letters. The message suggests an uppercased form of the key, for example
`database_url should use CONSTANT_CASE: DATABASE_URL`.

The suggestion is the key with its ASCII lowercase letters uppercased, and
nothing else — word separators are never inserted, so `apiToken` suggests
`APITOKEN` rather than `API_TOKEN`. Where a key needs new underscores to read
well, treat the suggestion as a starting point rather than the final name.

> [!NOTE]
> Keys with structural problems such as hyphens, dots, spaces or leading
> digits are reported by `invalid-key-name` instead. constant-case only
> reports otherwise-valid keys whose sole deviation is lowercase letters.

## Why it matters

CONSTANT_CASE is the widely accepted convention for environment variable
names and keeps keys consistent across files, shells and deployment tools.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking key names for lowercase ASCII letters.

## Fix

Rename the key to its CONSTANT_CASE form. There is no automatic fix: renaming
a key can break external references, so `vaar lint --fix` leaves these
findings untouched.

## Bad example

```dotenv
database_url=value
```

## Good example

```dotenv
DATABASE_URL=value
```
