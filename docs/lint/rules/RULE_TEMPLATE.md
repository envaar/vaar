<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# <rule-name>

One short sentence that states what the rule checks.

## What it catches

Describe the exact pattern or condition the rule flags.

## Why it matters

Explain the user-facing impact, failure mode or drift this rule prevents.

## Evidence level

Choose the weakest accurate evidence category for the rule:

- `Deterministic`: Vaar can prove the issue directly from the .env file contents alone.
- `Contextual`: Vaar can support the claim based on the repo context and codebase/project contents.
- `Heuristic`: Vaar sees something suspicious or worth reviewing based on heuristic patterns.
- `External`: Vaar uses another tool, API or integration to confirm the issue.

Additionally, justify why the rule falls under the chosen evidence category.

If the rule is not deterministic, state so plainly with reasoning and use words like `possible`, `warning` or `review required` when you describe the
result.

## Fix

Describe the simplest fix. If the rule has no automatic fix, say so.

If the rule is also fixed automatically by `vaar lint --fix`, add a short note in a markdown callout, for example:

> [!NOTE]
> This is also fixed automatically by running `vaar lint --fix`.

## Bad example

```dotenv
# replace with a real bad example
```

## Good example

```dotenv
# replace with a real good example
```
