# Verification Checklist — release/provenance-linux-packages

## Focused release gates

- [ ] `goreleaser check`
- [ ] `goreleaser release --snapshot --clean`
- [ ] `./scripts/verify-release-assets.sh dist`
- [ ] `./scripts/create-release-trust-fixtures.sh dist`
- [ ] `ALLOW_UNSIGNED_TRUST_FIXTURES=1 REQUIRE_TRUST_EVIDENCE=1 ./scripts/verify-release-assets.sh dist`
- [ ] `./scripts/test-linux-packages.sh dist` when Docker is available

## Workflow/security review

- [ ] Pull request path has only `contents: read` and does not request `id-token`, `attestations`, `artifact-metadata`, or `contents: write`.
- [ ] Release-only job isolates `contents: write`, `id-token: write`, `attestations: write`, and `artifact-metadata: write`.
- [ ] Release ordering is package/native final bytes → checksums → GitHub attestations/Cosign bundles → verification → upload.
- [ ] Upload list includes only verified release subjects and their Cosign bundles.
- [ ] No persistent signing private key, repository signing key, Apple/Windows signing, APT repo, or RPM repo claims are introduced.

## Repository gates

- [ ] `gofmt -w cmd internal` (not expected to change Go formatting)
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify` or recorded blocker if tool/runtime availability prevents full local execution

## Documentation checks

- [ ] `docs/release-verification.md` documents exact repository/workflow/ref-constrained commands.
- [ ] Documentation distinguishes standalone `.deb`/`.rpm` packages from signed APT/RPM repositories.
- [ ] Documentation states checksums alone provide integrity, not publisher identity.
