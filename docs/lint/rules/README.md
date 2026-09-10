<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Rule Reference

This directory contains documentation for all rules followed by Vaar during linting. Every rule document contains details about the rule, its importance and what it catches, the fix and both positive and negative example cases.

The list of all rules implemented is given in [Rule Catalog](#rule-catalog)

Vaar separates all rule documentation pages into folders based on evidence-level:

- `deterministic/` for exact findings proven from local files, schemas, configuration, Git state or other directly inspectable local evidence.
- `contextual/` for exact or bounded findings produced through supported repository, language, framework, CI, container or project analysis.
- `heuristic/` for candidate findings (warnings) based on patterns, naming, entropy, inference, scoring or commonly accepted practices where manual review is required.
- `external/` for findings verified through external tools, APIs, provider state, runtime instrumentation and so on.

See [Evidence Categories Reference](#evidence-levels) below.

## Evidence Levels

Use the accurate matching evidence level for each rule page:

| Level           | What it means                                                                                      | Example 1                                                                 | Example 2                                                                                                                         |
| --------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `Deterministic` | Obvious issues flagged deterministically                                                           | `duplicate-key` finds the same key twice in one file.                     | `trailing-whitespace` finds spaces or tabs at the end of a line.                                                                  |
| `Contextual`    | Issues flagged by deriving context using repo and codebase.                                        | A variable appears in source code but is missing from `.env.development`. | A key used in code never appears in the repo's `.env.example` file.                                                               |
| `Heuristic`     | Issues flagged based on mismatch/deviation from patterns, common code practices and standards etc. | A public file like `.env.example` contains a secret-like string.          | A repo file may contain a secret like string which is not present even in the `.env.*` file and may cause sensitive data leakage. |
| `External`      | Issues flagged by verification through another tool, API or integration.                           | A Cloud API confirms a required secret is absent.                         | An external scanner reports the secret as exposed.                                                                                |

## Rule Catalog

| Rule                                                          | Evidence level | Purpose                                                     |
| ------------------------------------------------------------- | -------------- | ----------------------------------------------------------- |
| [duplicate-key](./deterministic/duplicate-key.md)             | Deterministic  | Flags the same key appearing more than once in one file.    |
| [incorrect-delimiter](./deterministic/incorrect-delimiter.md) | Deterministic  | Flags `:` being used where `=` is expected.                 |
| [key-without-value](./deterministic/key-without-value.md)     | Deterministic  | Flags a key that has no assigned value.                     |
| [value-without-key](./deterministic/value-without-key.md)     | Deterministic  | Flags a value-like line that does not have a valid key.     |
| [leading-character](./deterministic/leading-character.md)     | Deterministic  | Flags leading whitespace before content.                    |
| [quote-character](./deterministic/quote-character.md)         | Deterministic  | Flags unbalanced quote usage.                               |
| [value-without-quotes](./deterministic/value-without-quotes.md) | Deterministic  | Flags unquoted values containing whitespace.                |
| [substitution-key](./deterministic/substitution-key.md)       | Deterministic  | Flags malformed env-var substitution syntax inside values.  |
| [space-character](./deterministic/space-character.md)         | Deterministic  | Flags spaces around the key, delimiter or value.            |
| [trailing-whitespace](./deterministic/trailing-whitespace.md) | Deterministic  | Flags trailing spaces or tabs.                              |
| [ending-blank-line](./deterministic/ending-blank-line.md)     | Deterministic  | Flags a file ending that does not match the current policy. |
| [extra-blank-line](./deterministic/extra-blank-line.md)       | Deterministic  | Flags repeated blank lines.                                 |
| [invalid-key-name](./deterministic/invalid-key-name.md)       | Deterministic  | Flags non-portable environment variable names.              |
| [constant-case](./deterministic/constant-case.md)             | Deterministic  | Flags portable keys that are not CONSTANT_CASE.             |
| [bom-character](./deterministic/bom-character.md)             | Deterministic  | Flags a UTF-8 BOM at the start of a file.                   |
| [line-ending](./deterministic/line-ending.md)                 | Deterministic  | Flags mixed CRLF and LF line endings.                       |

> [!NOTE]
> If you wish to request the addition of a new linting rule for Vaar, please open a new [Issue](https://github.com/envaar/vaar/issues) using the Rule Request template.

## Adding a rule page

When you add a new rule reference:

- choose the matching evidence-level directory
- start from [RULE_TEMPLATE.md](./RULE_TEMPLATE.md)
- add `docs/lint/rules/<evidence-level>/<slug>.md`
- add a row to the catalog above

> [!NOTE]
> Ensure to keep the page aligned with the rule template. Also ensure that the evidence level chosen is accurate. If a rule is only an indicator, do not phrase it as a deterministic check. Choose the evidence level that matches the strongest claim Vaar can support.
