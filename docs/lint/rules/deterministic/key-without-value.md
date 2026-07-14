<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# key-without-value

Flags a key that has no value assigned.

## What it catches

Lines where the key is present but the value is missing or empty.

## Why it matters

An incomplete assignment can break local setup or CI.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking for a key token with no assigned value.

## Fix

Provide a value or remove the line.

## Bad example

```dotenv
KEY=
```

## Good example

```dotenv
KEY=value
```
