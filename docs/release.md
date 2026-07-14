<!-- Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0 -->

# Releasing

This page is a guide for creating a production release of Vaar.

For the full workflow, repository settings, required secrets, release-day checks and cleanup steps, read the [Release Process Deep Dive](./release-process-deep-dive.md).

## How Releases Work

Vaar uses two tools to ship a release:

- `release-please` manages version bumps, release pull requests, changelog updates, release notes, tags and GitHub Releases.
- GoReleaser packages the release archives and the checksum manifest.

The normal release flow is:

1. `release-please` watches pushes to `main` and opens or updates a release pull request.
2. The release pull request updates `CHANGELOG.md` and `.release-please-manifest.json`.
3. Merging the release pull request creates the version tag and the GitHub Release.
4. The `release-artifacts` workflow runs after the release is published.
5. GoReleaser builds the archives and uploads them to the existing GitHub Release.

> [!NOTE]
> Release PR merges require maintainer approval.

## What Ships

Each release publishes the following platform archives and checksum manifest:

- `vaar_linux_amd64.tar.gz`
- `vaar_linux_arm64.tar.gz`
- `vaar_darwin_amd64.tar.gz`
- `vaar_darwin_arm64.tar.gz`
- `vaar_windows_amd64.zip`
- `vaar_windows_arm64.zip`
- `vaar_checksums.txt`

> [!NOTE]
> Checksums are the currently used integrity mechanism for all releases. Detached signatures and provenance attestations are not part of the current release line.

## Validate a Release Locally

Use the snapshot build when you want to exercise the packaging path without publishing anything:

```sh
make snapshot
```

Use the real release path from an exact version tag when you want to test the full release build:

```sh
git tag v0.1.0
make release
```

`make release` fails unless the current commit is tagged with an exact version such as `v0.1.0`.

## Release Readiness

Before a release pull request merges, confirm:

- the PR title follows the conventional commit format used by release-please, including accepted types such as `feat`, `fix` and `hotfix`
- the changelog entry(ies) looks correct
- the manifest version is the version you expect
- the release pull request has maintainer approval

## After the Release

After GitHub publishes the release, confirm:

- the `release-artifacts` workflow ran successfully
- the GitHub Release has the expected archives
- `vaar_checksums.txt` matches the uploaded artifacts
- an extracted binary reports the shipped version with `vaar --version`
