# Verification Checklist — release/provenance-linux-packages

## Focused release gates

- [x] Review fix focused check: shell syntax, expected release asset helper output, shellcheck, and actionlint passed.
- [x] `go run github.com/goreleaser/goreleaser/v2@latest check` — passed.
- [x] `SOURCE_DATE_EPOCH=$(git log -1 --format=%ct) go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean` — passed.
- [x] Reproducibility check: two snapshot builds with the same `SOURCE_DATE_EPOCH` produced identical hashes for all 11 copied release files (`checksums.txt` plus 10 archives/packages).
- [x] Docker Ubuntu `./scripts/verify-release-assets.sh dist` with `rpm` installed — passed, `verified 10 release assets`.
- [x] Docker Ubuntu `./scripts/create-release-trust-fixtures.sh dist` — passed, `11 subjects`.
- [x] Docker Ubuntu `ALLOW_UNSIGNED_TRUST_FIXTURES=1 REQUIRE_TRUST_EVIDENCE=1 ./scripts/verify-release-assets.sh dist` — passed.
- [x] `./scripts/test-linux-packages.sh dist` — passed before review fix on local Docker architecture with Ubuntu and Fedora containers.
- [x] `shellcheck scripts/verify-release-assets.sh scripts/create-release-trust-fixtures.sh scripts/test-linux-packages.sh` — passed.

## Workflow/security review

- [x] Pull request path has only top-level `contents: read` and does not request `id-token`, `attestations`, `artifact-metadata`, or `contents: write`.
- [x] Release-only job isolates `contents: write`, `id-token: write`, `attestations: write`, and `artifact-metadata: write`.
- [x] Release ordering is package/native final bytes → checksums → GitHub attestations/Cosign bundles → verification → upload.
- [x] Upload list includes only verified release subjects and their Cosign bundles, skips duplicate complete PM asset sets, refuses partial/mismatched pre-existing PM assets, and does not use `--clobber`.
- [x] No persistent signing private key, repository signing key, Apple/Windows signing, APT repo, or RPM repo claims are introduced.
- [x] `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/release.yml` — passed.

## Repository gates

- [x] `gofmt -w cmd internal` — passed/no Go diffs.
- [x] `go vet ./...` — passed.
- [x] `go test ./...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `make verify` — passed.

## PR #560 CI repair gates

- [x] `scripts/gsd doctor` — passed.
- [x] `scripts/gsd prompt programming-loop init --phase pr-560-ci-repair --dry-run` — unavailable (`unknown GSD command: programming-loop`); continued with the recorded manual-GSD fallback.
- [x] `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — red, GO-2026-6061 through `google.golang.org/grpc v1.79.3`.
- [x] `PR_TITLE=$(gh pr view 560 --json title -q .title) PR_BODY=$(gh pr view 560 --json body -q .body) go run ./cmd/prissueguard` — red, PR body has explicit issue-first wording but no `Closes`/`Refs` token.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard` — passed.
- [x] `PR_TITLE=$(gh pr view 560 --json title -q .title) PR_BODY=$(gh pr view 560 --json body -q .body) go run ./cmd/prissueguard` — passed, `issueguard: ok (3 linked issues)`.
- [x] `go test ./...` — passed.
- [x] `go vet ./...` — passed.
- [x] `go mod verify` — passed, all modules verified.
- [x] `go build ./cmd/pm` — passed.
- [x] `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — passed, no vulnerabilities found.
- [ ] `make verify` — attempted; stopped at `tidy-check` because this uncommitted PR repair intentionally changes `go.mod`/`go.sum`. Its remaining Go constituents were run directly above.

## PR #560 Snyk repair gates

- [x] `npm audit --omit=dev --package-lock-only --json` in `website/` — red, high vulnerabilities in `next`, `js-yaml`, `postcss`, and `sharp`.
- [x] `npx -y pnpm@11.7.0 install --lockfile-only --ignore-scripts` in `website/` — passed, lockfile passed pnpm supply-chain policy verification.
- [x] `npm install --package-lock-only --ignore-scripts` in `website/` — passed, reported `found 0 vulnerabilities`.
- [x] `npm audit --omit=dev --package-lock-only --json` in `website/` — passed, zero production vulnerabilities.
- [x] `npx -y pnpm@11.7.0 audit --prod` in `website/` — passed, no known production vulnerabilities.
- [x] `npx -y pnpm@11.7.0 install --frozen-lockfile --ignore-scripts` in `website/` — passed, lockfile is reproducible with the CI pnpm version.
- [x] `npx -y pnpm@11.7.0 run postinstall` then `npx -y pnpm@11.7.0 run typecheck` in `website/` — passed after generating the same Fumadocs source module that CI creates during install.

## Documentation checks

- [x] `docs/release-verification.md` documents repository/workflow/ref-constrained commands and digest/checksum verification.
- [x] Documentation distinguishes standalone `.deb`/`.rpm` packages from signed APT/RPM repositories.
- [x] Documentation states checksums alone provide integrity, not publisher identity.
