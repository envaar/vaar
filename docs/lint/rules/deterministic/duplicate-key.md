<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# duplicate-key

Flags a key that appears more than once in the same dotenv file.

## What it catches

Repeated assignments of the same environment variable in one file.

## Why it matters

The later value can silently suppress the earlier one.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by detecting repeated assignments of the same key in one file.

## Fix

Keep one definition and remove the duplicate entry.

## Bad example

```dotenv
KEY=value
KEY=other
```

## Good example

```dotenv
KEY=value
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=duplicate-key
```

Bad Example Produces:

```text
error duplicate-key .env:2 KEY is defined more than once
```
