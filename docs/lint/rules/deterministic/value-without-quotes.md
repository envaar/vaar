# value-without-quotes

Reports unquoted values containing whitespace.

## Bad

```dotenv
FOO=BAR BAZ
WELCOME_MESSAGE=Hello world
DATABASE_LABEL=Primary database
```

## Good

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
.env:1 value-without-quotes: value containing whitespace should be enclosed in quotes
```

## Evidence level

Deterministic