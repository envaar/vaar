<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# value-without-key

Flags a value-like line that does not have a valid key.

## What it catches

Malformed lines that start with a value or delimiter before any key.

## Why it matters

The file is not representing a valid environment variable assignment.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by detecting value-like lines that do not have a valid key.

## Fix

Add the missing key or remove the malformed line.

## Bad example

```dotenv
just some text
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=value-without-key
```

Bad Example Produces:

```text
error value-without-key .env:1 value appears without a valid key
```
