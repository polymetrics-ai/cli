# Plan — release/provenance-linux-packages

**Issues:** Closes polymetrics-ai/cli#551 and polymetrics-ai/cli#552; refs polymetrics-ai/cli#550.
**Branch:** `release/provenance-linux-packages`
**PR title:** `build(release): add provenance and Linux packages`

## GSD command path

- `scripts/gsd doctor` passed in this worktree.
- `scripts/gsd list` showed no `programming-loop` command.
- `scripts/gsd prompt programming-loop init --phase release-provenance-linux-packages --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Fallback: manual GSD programming loop per `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` with plan, TDD ledger, verification checklist, red validation, implementation, verification, and committed green slice.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-continuous-integration`
- `golang-security`
- `golang-dependency-management`
- `golang-testing`
- `golang-documentation`
- `golang-lint`
- `context-mode`

## Scope

1. Add GoReleaser v2/nFPM Linux `.deb` and `.rpm` packages for Linux `amd64` and `arm64`.
2. Add package metadata grounded in repository files: package name `pm`, homepage `https://cli.polymetrics.ai`, vendor/maintainer `Polymetrics AI`, license `AGPL-3.0-only`, binary `/usr/bin/pm`, and license/notice docs under `/usr/share/doc/pm/`.
3. Extend release verification to enforce archives, native packages, checksums, package contents, package architecture metadata, and optional trust evidence.
4. Add release-only keyless trust evidence for future releases: GitHub artifact attestations and Cosign keyless blob bundles for all final release archives/packages and `checksums.txt`.
5. Keep PR validation offline: verify deterministic asset structure, package inspection, package install/remove in clean Linux containers where feasible, and fixture-mode trust evidence without publishing or contacting production signing services.
6. Add narrowly scoped release verification documentation for checksums, Cosign bundles, GitHub attestations, and standalone package trust boundaries.

## Non-goals

- Do not alter published `v0.1.0` assets.
- Do not complete Apple Developer ID/notarization, Windows Authenticode, WinGet, or signed APT/RPM repositories.
- Do not add repository signing keys, APT repository metadata, RPM repository metadata, or persistent signing private keys.
- Do not add Go module dependencies.
- Do not change unrelated connector docs, generated website data, or CLI command behavior.

## Implementation slices

1. **Red validation:** run current release verification expectations and note that packages/trust evidence are absent from the baseline.
2. **Package slice:** update `.goreleaser.yaml`, `scripts/verify-release-assets.sh`, and CI package checks for nFPM packages and unprivileged package inspection.
3. **Trust slice:** update `.github/workflows/release.yml` for release-only OIDC attest/sign/verify/upload ordering and offline fixture validation in PRs.
4. **Docs slice:** add narrow `docs/release-verification.md` and link only where already release-verification text exists.
5. **Verification:** GoReleaser snapshot build, release verification (structure and fixture trust), package inspection/install tests, workflow syntax/security review, relevant Go checks, and shell checks.

## Execution decision

`local_critical_path`: this is one focused release pipeline slice touching shared release workflow/config/scripts/docs. No mutating subagent was spawned because this disposable worktree is the isolated worker context and the edits are tightly coupled.

## Review fix slice

1. Replace count-only existing-release skip logic with exact asset-set comparison, release asset download, checksum validation, Cosign bundle verification, and GitHub attestation verification before skipping upload.
2. Extend Linux package install coverage to run both amd64 and arm64 packages in architecture-matched Docker containers, using QEMU on GitHub-hosted amd64 runners for arm64.
3. Run one focused verification pass over the changed release workflow and shell scripts after all fixes are applied.

## CI repair slice for PR #560

1. Resolve `govulncheck` GO-2026-6061 by upgrading the existing indirect `google.golang.org/grpc` requirement from `v1.79.3` to the fixed floor `v1.82.1`, then run `go mod tidy` and the CI vulnerability command with `GOTOOLCHAIN=go1.25.12`.
2. Fix `require-linked-issue` by adding a focused red test for PR #560's issue-first intent wording, then make `cmd/prissueguard` accept explicit delivery/parent issue references such as `fully delivering issues #551 and #552` and `parent polymetrics-ai/cli#550` while keeping vague references rejected.
3. Verify with focused issueguard tests, `go test ./cmd/prissueguard ./internal/coordination/issueguard`, `go test ./...`, and the exact `govulncheck` CI command.
