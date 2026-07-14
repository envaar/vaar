<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Contributing

Thank you for helping us improve Vaar.

## Before you start

Please read through the documentation given below to understand this project and its goals.

- [README.md](./README.md)
- [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)
- [SECURITY.md](./SECURITY.md) for sensitive reports or redaction concerns
- [System Overview](./docs/system-overview.md) for package boundaries and runtime flow

> [!NOTE]
> For any change, please open an [Issue](https://github.com/envaar/vaar/issues) or [Discussion](https://github.com/envaar/vaar/discussions) first and get it approved by a [Maintainer](mailto:core@envaar.dev).

## Development Workflow

- Create a fork of this repository on your account.
- Keep pull requests focused on a single [issue](https://github.com/envaar/vaar/issues).
- Add or update tests when behavior changes.
- Update [docs](/docs/) when command behavior, output or contributor expectations change.
- Run `gofmt -w <touched Go files>` on changed Go files and use `make lint` and `make vet` to check formatting.
- Run `make test` before opening a pull request.
- Run `make build` to confirm the binary still builds.
- Raise a PR and request review from a [Maintainer](mailto:core@envaar.dev).

> [!NOTE]
> Run `make bench` if the change affects file walking, parsing or rule runtime

## Important contributor commands

- `make lint` checks formatting and runs the CLI smoke test, which catches style drift and broken user-facing behavior early.
- `make test` runs the full Go test suite, which verifies that code changes did not break the project.
- `make vet` runs `go vet`, which helps catch suspicious patterns and correctness issues the compiler will not flag.
- `make build` builds the CLI binary, which confirms the project still compiles into a working executable.

## Pre-commit guidance

The lightweight default is the repository hook in [scripts/pre-commit](./scripts/pre-commit).

Install it into your clone using:

```bash
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

The hook runs `make lint`, so formatting and the CLI smoke test fail before the commit lands.

If your team already uses the `pre-commit` framework, the same behavior maps to a local hook as follows:

```yaml
repos:
  - repo: local
    hooks:
      - id: vaar-lint
        name: vaar lint
        entry: scripts/pre-commit
        language: system
        pass_filenames: false
```

Use `make test` before opening a pull request when you need a broader check than the hook.

## Commit messages and PR titles

Vaar uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

> [!NOTE]
> release-please owns release versioning, release pull requests, generated release notes and versioned changelog sections. Do not hand-edit versioned changelog sections outside the release PR.

- Use `<type>: <description>` for commits and PR titles.
- Keep descriptions short, imperative and specific.
- Use `feat` for new user-visible behavior.
- Use `fix` for bug fixes.
- Use `docs`, `test`, `refactor`, `chore`, `ci`, `build` or `perf` when necessary.
- Use `!` only when a [Maintainer](mailto:core@envaar.dev) explicitly approves a breaking change.
- If you are unsure about your change or need any clarifications, please discuss in the related [Issue](https://github.com/envaar/vaar/issues) or [Discussion](https://github.com/envaar/vaar/discussions).

> [!NOTE]
> The PR title becomes the squash-merge commit message, so the title must follow the same format.

## Pull Requests

- Use the provided pull request template.
- Explain what changed and why.
- Call out any output changes, parser changes or rule semantics changes.
- Follow the commit message rules above.
- Ensure redaction of actual secret values in examples, tests or logs in the code itself.

## Code style

- Keep Go code formatted with `gofmt`.
- Prefer clear, explicit names for rules, flags and output messages.
- Please refer to already created rules and their respective documentation files at [/internal/lint/rules](/internal/lint/rules/) and [/docs/lint/rules](/docs/lint/rules/) respectively to better understand the project's coding standards and contribution quality while introducing new rules.

## Rule requests

> [!NOTE]
> If you wish to request the addition of a new linting rule for Vaar, please open a new [Issue](https://github.com/envaar/vaar/issues) using the Rule Request template.

While proposing a new lint rule, include:

- the problem you want Vaar to catch
- a bad example that should be reported
- a good example that should be allowed
- the expected output format
- the weakest accurate evidence level: deterministic, contextual, heuristic or external

If you are changing an existing rule, update the matching page in
`docs/lint/rules/deterministic/` so it keeps the same structure as
[docs/lint/rules/RULE_TEMPLATE.md](./docs/lint/rules/RULE_TEMPLATE.md):

- what it catches
- why it matters
- evidence level
- fix
- bad example
- good example

## Adding a New Rule

For a new rule, update the rule implementation, its tests and the matching docs together so the everything stays updated.

Create/Modify these files:

- `internal/lint/rules/<evidence-category>/<slug>.go`
- `internal/lint/rules/all.go`
- `internal/lint/rules/compat.go`
- `internal/lint/rules/<evidence-category>/<slug>_test.go`
- `docs/lint/rules/<evidence-category>/<slug>.md`
- `docs/lint/rules/README.md`

The CLI completions pick up new rule names automatically from `internal/lint/rules/all.go`, so there is no separate completion wiring to add.

If the rule changes user-visible command output or rule selection behavior also update the matching tests and docs such as `README.md`, `docs/usage.md`, `docs/lint/README.md`, `docs/lint/rules/README.md`.
