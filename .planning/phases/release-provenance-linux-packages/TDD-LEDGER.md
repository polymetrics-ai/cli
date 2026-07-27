# TDD Ledger — release/provenance-linux-packages

| Step | Type | Command / artifact | Expected result | Status |
|---|---|---|---|---|
| 1 | Red validation | `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean`; `./scripts/verify-release-assets.sh dist` | Baseline verifier passed with only six archives and no `.deb`, `.rpm`, or `.sigstore.json` evidence, proving the release gate did not yet enforce issues #551/#552 | Red captured |
| 2 | Green validation | `SOURCE_DATE_EPOCH=$(git log -1 --format=%ct) goreleaser release --snapshot --clean`; Docker Ubuntu `./scripts/verify-release-assets.sh dist` with `rpm` installed | Expected archives, `.deb`, `.rpm`, package metadata/contents, and checksums pass (`verified 10 release assets`) | Green |
| 3 | Green validation | Docker Ubuntu `./scripts/create-release-trust-fixtures.sh dist`; `ALLOW_UNSIGNED_TRUST_FIXTURES=1 REQUIRE_TRUST_EVIDENCE=1 ./scripts/verify-release-assets.sh dist` | Offline fixture mode checks subject names/digests without signing services (`11 subjects`) | Green |
| 4 | Green validation | `./scripts/test-linux-packages.sh dist` | Clean Ubuntu and Fedora containers install, exercise, reinstall/upgrade, and remove packages for each architecture under test | Green |
| 5 | Release-only validation | `.github/workflows/release.yml` trust step; `actionlint` | GitHub attestations and Cosign bundles are generated after final checksums and verified before upload with release-only OIDC/attestation permissions | Green by static validation; production attestation minting only runs on release |
| 6 | Broader checks | `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, `make verify` | Repository remains green; release changes do not add Go dependencies | Green |
| 7 | Review red validation | Review finding audit against `.github/workflows/release.yml` and `scripts/test-linux-packages.sh` | Existing-release skip is count-only and package install tests are host-architecture-only | Red captured |
| 8 | Review green validation | `bash -n ...`; `./scripts/verify-release-assets.sh --release-version 1.2.3 --print-expected-release-assets`; `shellcheck ...`; `actionlint .github/workflows/release.yml` | Existing-release skip proves exact assets/checksums/Cosign/GitHub attestations; package install tests cover amd64 and arm64 containers | Green |

## Notes

- Snapshot PR validation must not mint production attestations or publish release assets.
- Fixture trust evidence is explicitly unsigned and accepted only when `ALLOW_UNSIGNED_TRUST_FIXTURES=1` is set.
- Release upload must fail if any required asset, checksum entry, Cosign bundle, GitHub attestation, or package verification is missing or mismatched.
