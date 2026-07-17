<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# value-without-quotes

Flags unquoted values containing whitespace.

## What it catches

Values that contain spaces or tabs but are not enclosed in matching single or double quotes.

## Why it matters

Unquoted whitespace can be parsed inconsistently across dotenv tooling and can make the intended value ambiguous.

## Evidence level

- `Deterministic`: Vaar can prove this from the parsed file contents alone by checking whether an assigned value contains whitespace outside matching quotes.

## Fix

Wrap the full value in matching single or double quotes.

Automatic fixing is not currently provided because Vaar must avoid changing interpolation behavior, escapes, embedded quotes or inline comment parsing.

## Bad example

```dotenv
FOO=BAR BAZ
WELCOME_MESSAGE=Hello world
DATABASE_LABEL=Primary database
```

## Good example

```dotenv
FOO="BAR BAZ"
WELCOME_MESSAGE='Hello world'
DATABASE_LABEL="Primary database"
FOO=BAR
PORT=3000
DEBUG=true
```

## Example output

```text
error value-without-quotes .env:1 value containing whitespace should be enclosed in quotes
```
