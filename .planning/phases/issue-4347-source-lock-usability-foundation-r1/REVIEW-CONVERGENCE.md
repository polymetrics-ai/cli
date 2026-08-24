# Review convergence — issue 4347 source-lock usability

## Freeze

- PR: #4350
- Branch: `fm/cli-source-lock-usability-foundation-r1`
- Base SHA: `060bb7864e3419e09ab10e000bb14ac1ea3724ec` (`origin/main`)
- Merge base: `060bb7864e3419e09ab10e000bb14ac1ea3724ec`
- Review SHA: `f92a16fe6f16772503676af276242100144e6ed9`
- Generated-file status: `go run ./cmd/connectorgen surface-sync --check` passed at the freeze. The tracked GitHub retained-artifact manifest is a provenance output updated by the explicit retain command; no generated-surface drift is present.

The source tree is read-only through discovery. Review is restricted to this
immutable SHA and the 13 changed files below.

1. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/CONTEXT.md`
2. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/DISCUSSION-LOG.md`
3. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/PLAN.md`
4. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/REVIEW.md`
5. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/RUN-STATE.json`
6. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/TDD-LEDGER.md`
7. `.planning/phases/issue-4347-source-lock-usability-foundation-r1/VERIFICATION.md`
8. `cmd/connectorgen/main.go`
9. `cmd/connectorgen/sourceartifact.go`
10. `cmd/connectorgen/sourceartifact_test.go`
11. `cmd/connectorgen/sourceimport.go`
12. `docs/migration/conventions.md`
13. `internal/connectors/defs/github/sources/github-retained-artifacts.json`

## Lens ledgers

| Lens | Status | Evidence / invariant |
| --- | --- | --- |
| Architecture / data flow | Complete | `loadConnectorSourceRetainLock` reads only connector-owned operation/parity locks and extracts artifacts; `validateSourceRetainArtifact` checks only URL, identity, digest, and size. `validateSourceImportArtifact` layers form-pin validation on top, and import still enters its inventory checks. `TestSourceRetainRetainsV3ArtifactWithoutImportTimeFormInventory` proves retain succeeds while `parseSourceImportLock` rejects that incomplete lock. |
| Happy / bad / edge | Complete | `TestSourceRetainWritesVerifiedMachineReadableArtifactWithoutChangingLock` proves a byte-exact artifact remains byte-exact after offline read. `TestSourceRetainVerifiesReorderedJSONByCanonicalIdentity` proves unequal serialisations with equal parsed JSON retain and verify through `canonical_json`. Mismatch, wrong-source, and BOT-BLOCK tests prove no invalid response reaches a write. |
| State machine / concurrency | Complete | A retain invocation fetches and validates every payload before writing. Artifacts are content-addressed by fetched-byte digest, staged/synced, and published through a no-replacement hard link; manifests are staged/synced and renamed. Existing directories/files and imported artifacts reject symlinks, so a partial or redirected file cannot become a valid retained input. |
| Security / secret taint | Complete | The connector lock, not CLI input, supplies the URL. Retain uses the existing public-address policy and bounded identity-query validator; the HTTP client refuses redirects. No credential, generic URL, request body, or secret-bearing input was added. `TestSourceRetainAllowsLockedParityQuery` confirms only the connector-owned fixed query is fetched. |
| Retry / rate limit / resume / idempotency | Complete | Reruns verify the pre-existing fetched-byte digest and selected lock identity before retaining it; a different existing file is refused rather than replaced. Manifest records are keyed by lock identity plus fetched-byte digest, allow legacy enrichment only for absent fields, and preserve explicit lock identity/provenance. No retry loop was introduced; transport timeout remains bounded by the existing client. |
| Output integrity | Complete | `validateSourceImportArtifactBytes` selects byte identity (exact size/SHA-256) or canonical JSON digest explicitly. Retain records selected identity plus `retained_sha256`/`retained_bytes`, while the offline reader verifies raw provenance first and selected lock identity second. Thus canonical-equivalent bytes are traceable without treating raw byte variation as drift; a byte lock still produces the legacy exact-byte mismatch diagnostic. |
| Declaration reachability / closed surface | Complete | Only `<connector>-operation-source-lock.json` and `<connector>-parity-source-lock.json` are read, both regular non-symlink files. Retention can preserve v3 document artifacts and parity artifacts without making them importable; source import keeps strict operation/inventory/form requirements. Tests cover operation, v3 incomplete-form, rendered/bundle, parity, and fixed-query lock variants. |
| CLI / app parity | Complete | `source-retain --help` and `docs/migration/conventions.md` state `byte` versus `canonical_json`, wrong-source, BOT-BLOCK, and no silent re-pin behavior; `TestSourceRetainHelpAndMigrationDocumentationDescribeIdentityAndWrongSource` checks that contract. This is a connectorgen maintenance command only, so no `pm` command, website, or app surface changed. |
| Provider semantics | Complete | Redirects are typed wrong-source at `CheckRedirect`; non-200 responses are wrong-source except HTTP 403, which is BOT-BLOCK. A login-wall HTML response or response under one eighth of the locked size is classified before identity comparison. `TestSourceRetainDoesNotMisclassifyExpectedHTMLWithLoginText` prevents false positives for a legitimate full-size HTML reference, while the wrong-source and BOT-BLOCK tests assert neither verdict is drift. |
| Tests / evidence | Complete | Red/green evidence is recorded in `TDD-LEDGER.md`; focused `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain'`, targeted retained-reader regression, `go vet ./cmd/connectorgen`, surface-sync, independent verify gates, and eight live byte-exact retains are recorded in `VERIFICATION.md`. The clean-main reproduction passed, locating the prior CI failure on this branch and its regression was restored at the frozen SHA. |

## Merged invariant ledger

| Invariant | Lenses merged | Result |
| --- | --- | --- |
| Retention never requires prior importability. | Architecture; declaration reachability; happy/bad/edge | The retain parser and HTTP entry point use retain-only validation, while import still requires stricter form/inventory facts. No bypass of import validation was found. |
| A response cannot be silently retained unless it satisfies the selected lock identity and its own recorded fetched-byte provenance. | Happy/bad/edge; state; output integrity; retry/idempotency | Classification happens before identity comparison; writes require selected identity, retain raw digest/length, and refuse replacement. The retained reader rechecks both identities. No silent-retention path was found. |
| A genuinely byte-exact lock remains byte-exact, and semantic JSON variation is not misreported as drift. | Happy/bad/edge; output integrity; tests | Default `byte` remains size plus SHA-256. `canonical_json` is opt-in and retains a separate fetched-byte record; the test exercises unequal serialisations and offline verification. No loss of byte-lock semantics was found. |
| Bad provider access is actionable and distinct from drift. | Provider semantics; security; CLI/app parity; tests | Redirect/login/undersize/404 classify as wrong-source and 403/TLS as BOT-BLOCK before drift validation; help/docs name the remedies. No error-page, login-wall, or redirect pinning path was found. |
| Retained files stay safe across repeated runs. | State; retry/idempotency; security | Non-symlink regular-file checks, sync-before-publish, no-replacement artifact links, and manifest atomic replacement preserve a valid/readable state. No race or partial-publication defect was found in the single-process retention flow. |

## Independent audit

The requested audit was executed in a fresh, read-only context against review
SHA `f92a16fe6f16772503676af276242100144e6ed9` and the merged ledger. The
available GSD adapter did not expose a callable `--harness claude` switch, so
the equivalent isolated audit was used; it made no file changes.

| Finding | Severity | Evidence | Required disposition |
| --- | --- | --- | --- |
| Uppercase lock SHA-256 cannot round-trip through retention. | Medium | `validateSourceRetainArtifact` accepts uppercase hexadecimal. Manifest construction lowercases `SHA256`, but the retained-reader record matcher compares the lock SHA-256 case-sensitively. The record index key is already lowercase, so the successful retain later reports missing matching provenance. | Add a retain-to-offline-read red regression using an uppercase lock digest; compare lock digest case-insensitively; re-run the exact final-SHA audit. |

No other blocker was found: retain-only validation, canonical/raw provenance
ordering, no-replacement storage, bad-source classification, and exact-byte
identity stayed intact. This finding starts the coordinated fix wave; any source
change creates a new review SHA.
