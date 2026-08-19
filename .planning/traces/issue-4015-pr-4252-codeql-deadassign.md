# PR #4252 CodeQL Dead-Assignment Fix

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `release/0.2.0-mvp`
- Merges into: `fm/cli-release-codeql-deadassign-r1 → release/0.2.0-mvp → main` via PR #4250
- Delivery: PR #4252 green and ready for integration into the release branch; no merge to `main`
- Working branch: `fm/cli-release-codeql-deadassign-r1`
- Task: Remove the CodeQL-reported dead assignment in `parseLegacySyncMode` without changing behavior

## GSD / TDD Evidence

- Lifecycle: `scripts/gsd prompt discuss-phase pr-4250-codeql-deadassign-r1` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, executed with the documented inline/manual GSD fallback for this two-line release fix.
- Required skills: `golang-how-to`, `golang-testing`, `golang-lint`, `golang-error-handling`, `golang-safety`, `golang-security`, and `golang-documentation` for this evidence follow-up.
- Red: CodeQL reported “Useless assignment to local variable” at `internal/app/sync_modes.go:154`. Inspection confirmed the first `SyncMode` returned by `publicSyncMode` was never read: both error paths return `SyncMode{}`, and the success path overwrote `mode` before returning it.
- Green: Discard the unused first result from `publicSyncMode` and declare `mode` where `definition.persistedLegacyMode()` produces the value. Control flow, errors, empty-string defaulting, and returned fields remain unchanged.
- Refactor: None beyond the discard-and-declare change.

## Verification

- `go test -count=1 -timeout 20m -run 'Test(ParseSyncModeSeparatesLegacyCompatibilityFromNewNativeAdmission|DedupedLegacyAliasesUseTypedContractsBeforeSourceIO|FreshLegacyConnectionCreationMarksAllPublicModes)$' ./internal/app` — pass (`4.521s`).
- `go test -count=1 -timeout 20m ./internal/app` — pass (`254.767s`).
- `go vet ./internal/app` — pass.
- `go build ./cmd/pm` — pass.
- `golangci-lint run --new-from-rev=HEAD ./internal/app` — pass with zero issues.
- `make tidy-check` plus the repository docs, smoke, agent-contract, connector generation/snapshot, boundary/canon, and release-workflow checks — pass.
- Manual code review found no behavior, security, or error-path regression.

## Preservation Proof

- Certification evidence files under `internal/connectors/certifications/evidence`: **998**.
- GitHub commands declared by `internal/connectors/defs/github/cli_surface.json`: **1571**.
- GitHub commands with `availability: implemented`: **1546**.
