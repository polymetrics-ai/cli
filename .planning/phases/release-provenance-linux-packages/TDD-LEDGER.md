# TDD Ledger — release/provenance-linux-packages

| Step | Type | Command / artifact | Expected result | Status |
|---|---|---|---|---|
| 1 | Red validation | `./scripts/verify-release-assets.sh dist` after current snapshot | Baseline verifier does not require Linux native packages or trust evidence; record gap before production edits | Planned |
| 2 | Green validation | `goreleaser release --snapshot --clean` then `./scripts/verify-release-assets.sh dist` | Expected archives, `.deb`, `.rpm`, package metadata/contents, and checksums pass | Planned |
| 3 | Green validation | `scripts/create-release-trust-fixtures.sh dist` then `ALLOW_UNSIGNED_TRUST_FIXTURES=1 REQUIRE_TRUST_EVIDENCE=1 ./scripts/verify-release-assets.sh dist` | Offline fixture mode checks subject names/digests without signing services | Planned |
| 4 | Green validation | `scripts/test-linux-packages.sh dist` | Clean Debian/Ubuntu and Fedora/RHEL-family containers install, exercise, reinstall/upgrade, and remove amd64 packages | Planned |
| 5 | Release-only validation | Release workflow trust step | GitHub attestations and Cosign bundles are generated after final checksums and verified before upload | Planned |
| 6 | Broader checks | `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, `make verify` as feasible | Repository remains green; release changes do not add Go dependencies | Planned |

## Notes

- Snapshot PR validation must not mint production attestations or publish release assets.
- Fixture trust evidence is explicitly unsigned and accepted only when `ALLOW_UNSIGNED_TRUST_FIXTURES=1` is set.
- Release upload must fail if any required asset, checksum entry, Cosign bundle, GitHub attestation, or package verification is missing or mismatched.
