<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# quote-character

Flags unbalanced quoting in a dotenv value.

## What it catches

Values that start with a quote but do not close cleanly.

## Why it matters

Broken quoting changes how the rest of the line is parsed.

## Evidence level

- `Deterministic`: Vaar can prove this from the file contents alone by checking whether a quoted value closes cleanly.

## Fix

Close the quote or remove it.

## Bad example

```dotenv
KEY="value
```

## Good example

```dotenv
KEY="value"
```

## Example output

Linting the bad example above, saved as `.env`:

```bash
vaar lint --only=quote-character
```

Bad Example Produces:

```text
error quote-character .env:1 value has unbalanced quotes
```
