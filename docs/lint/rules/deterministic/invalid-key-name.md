<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# invalid-key-name

Flags keys that are not portable environment variable names.

## What it catches

Keys that do not match the portable env-key pattern.

> [!NOTE]
> Portable env-key names start with a letter or `_`, then use only letters,
> digits or `_`.

## Why it matters

Portable key names work more reliably across shells, CI and deployment tools.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by matching key names against the portable environment-variable pattern.

## Fix

Rename the key to a portable form.

## Bad example

```dotenv
bad-key=value
```

## Good example

```dotenv
BAD_KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=invalid-key-name
```

Bad Example Produces:

```text
error invalid-key-name .env:1 bad-key is not a portable env key name
```
