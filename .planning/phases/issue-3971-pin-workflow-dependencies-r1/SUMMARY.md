# Summary: pin build dependencies for Scorecard alert #135

## Delivered

- Pinned every external GitHub Action in all audited workflows to its current full commit SHA with the original version retained in a trailing comment.
- Pinned `node:26-alpine` in `website/Dockerfile` and `postgres:17-alpine` in the website CI service to their current manifest-list digests.
- Added `scripts/tests/pinned-build-dependencies.sh` and wired it into `make release-workflow-check` so mutable action/image refs fail locally and in `make verify`.
- Created child issue [#3986](https://github.com/polymetrics-ai/cli/issues/3986) and attached it to GitHub parity parent #3971 after duplicate searches found no equivalent issue.
- After #3970 merged, rebased onto current `origin/main` and pinned its newly introduced `github-source-drift.yml` checkout and setup-node actions to the same captured `v7`/`v6` commits.

## Alert inventory

The final REST read found five open Scorecard `PinnedDependenciesID` alerts, all at the pre-change default-branch commit:

| Alert | Location |
| --- | --- |
| #135 | `.github/workflows/verify.yml:86` |
| #134 | `.github/workflows/verify.yml:81` |
| #133 | `.github/workflows/release.yml:480` |
| #132 | `.github/workflows/release.yml:422` |
| #131 | `.github/workflows/release.yml:415` |

All five are fixed by the same immutable-reference change. #135 is not closed yet because the branch has not been pushed or merged and Scorecard has not re-analyzed `main`.

## Verification

Red/green evidence, source-resolution proof, focused pinning/YAML checks, release-workflow checks, Dockerfile parsing, Go build/vet, lint, docs, smoke, contract, connector, and GSD workflow checks all passed. See `TDD-LEDGER.md`, `VERIFICATION.md`, and `REVIEW.md` for commands and results.
