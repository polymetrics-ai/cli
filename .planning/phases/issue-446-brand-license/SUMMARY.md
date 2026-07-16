# Summary: Issue 446

## Delivered

- Restored a single accessible `PmLogoMark` component after its deletion in
  commit `605b006e` (PR #29).
- The mark keeps `P` static, blinks only `M`, renders no underscore, and falls
  back to a stable `PM` for reduced-motion users.
- Replaced local logo implementations in the navbar, home sidebar, and footer.
- Changed the repository default from Elastic License 2.0 to the official GNU
  AGPL v3 text, identified as `AGPL-3.0-only`.
- Added an MIT boundary at `internal/connectors/defs/LICENSE`.
- Added `LICENSING.md` and aligned README, NOTICE, CONTRIBUTING, website copy,
  and automated review guidance with that path map.
- Added regression tests for both the brand and license contracts.

## Research Basis

- GNU's official AGPL v3 text and network-source terms.
- SPDX identifiers `AGPL-3.0-only` and `MIT`.
- Grafana's default-license plus path-specific `LICENSING.md` pattern.
- Repository provenance audit: one human author plus Dependabot on `main`.

## Verification

- Website unit tests: 70 passed.
- Website typecheck: passed.
- Website production build: passed.
- `go test ./...`, `go vet ./...`, and `go build ./cmd/pm`: passed.
- `make verify`: passed, including lint, smoke, docs, and 547 definitions with
  zero validator findings.
- Desktop/mobile/reduced-motion Playwright inspection: passed.

## Remaining Gate

This work is ready for automated review, but the license change requires
explicit repository-owner/legal approval before merge. No production deploy or
merge to `main` is part of this phase.
