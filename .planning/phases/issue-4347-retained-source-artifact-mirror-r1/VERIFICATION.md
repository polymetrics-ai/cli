# Issue #4347 — verification

## Result

The retained-reader foundation is locally green. Historic-byte preservation is
attempted first; **Firstmate** separately authorized an explicit re-pin only
after the per-artifact response is classified as a real provider document. An
error response remains irrecoverable and is never re-pinned.

## Executed verification

- `gofmt -w cmd/connectorgen/sourceartifact.go cmd/connectorgen/sourceimport.go cmd/connectorgen/sourceimport_test.go` — passed.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceRetain|TestSourceImportRetainedArtifact|TestSourceImportVersion2|TestSourceImportRejectsSymlinkedSourcesDirectoryEvenInsideConnectorBundle|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified)$' -count=1` — passed in 1.239s. Covers retained machine-readable import without a provider, missing/mismatched/symlinked copies, manifest tampering, rendered reference bytes, zip bundle bytes, raw v2 GraphQL pins, and Elasticsearch/Zoom recovery fixtures.
- `go test -timeout 20m ./cmd/connectorgen -count=1` — final pass in 156.093s.
- `go test -timeout 20m ./internal/cli -count=1` — final pass in 431.272s.
- `go vet ./...`, `go build ./cmd/connectorgen`, and `go build ./cmd/pm` — passed.
- `go run ./cmd/connectorgen source-import --help` and `git diff --check` — passed.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceRetain|TestSourceImportRetainedArtifact|TestSourceImportVersion2|TestSourceImportRejectsSymlinkedSourcesDirectoryEvenInsideConnectorBundle|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified)$' -count=1` — passed in 1.264s after the final reader/materializer additions.
- `go run ./cmd/connectorgen source-retain github --retrieved-at 2026-08-24T07:02:03Z --license 'GitHub REST API Description: MIT; GitHub Docs GraphQL schema: CC-BY-4.0' --terms 'GitHub REST API Description LICENSE and GitHub Docs Terms of Service; attribution and notices retained in provenance.'` — passed; wrote two lock-verified raw artifacts totaling 14,471,636 bytes and the tracked provenance manifest.
- `go run ./cmd/connectorgen source-import github --check` — final pass in 8.907s: `github, 1525 operation(s), 0 inbound event(s) verified`. The normal import command instantiated only the retained reader.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportPreservesFrozenGitHubArtifacts|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified|TestSourceRetain)' -count=1` — passed in 1.272s; the intentional GitHub lock re-pin has an updated frozen-byte snapshot.
- `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, and `make lint` — passed. The boundary report checked 317 files/553 connectors with zero findings; generator validation checked 553 connectors with zero findings, surface-sync filled/corrected zero fields, and release certification archive proof passed.
- Full `go test -timeout 20m ./...` / `make verify` were intentionally not run as one local command: repository instruction requires package-scoped checks in this per-command environment because the 550+ connector suite routinely exceeds the command window. CI remains the full-suite authority.

## Acceptance recovery evidence

- Before the write, `go run ./cmd/connectorgen source-import github --check` failed before any network path with `read retained artifact manifest: ... github-retained-artifacts.json: no such file or directory`. After the explicit `source-retain` maintenance command, the same hermetic check passes using the two tracked raw files.
- GitHub REST was preserved at `12,920,264` bytes / `80850d…d5b1d`. GitHub GraphQL was a Firstmate-authorized explicit re-pin from `1,546,421` / `c09aba…c249d` to the observed real GraphQL schema at `1,551,372` / `30347e…40c3d`; `REPIN-REPORT.json` records the old/new identities, HTTP 200 classification, retrieval date, and the complete pre-change tracked-lock Git blob.
- Read-only current-response measurement: all 27 Batch 8–10 nonzero source inputs plus GitHub are 454,667,903 bytes today versus 453,725,820 locked bytes. 351 Batch 8–10 documents differ from their pins; GitHub's GraphQL input also differs. They require the explicit Firstmate-authorized real-document classification and re-pin report before any retained-copy write.
- Zoom accounts: the historic dated URL returned HTTP 404 / 8,329 bytes. Elasticsearch: the historic URL returned HTTP 200 / 6,458,837 bytes versus the lock's 6,458,869 bytes. An exact SHA-256/length scan across every reachable Git blob found no matching historical copy for either artifact.
- The engine tests prove a correct retained copy would permit both cases to import without contacting the provider. They cannot truthfully claim recovery of the real historic Zoom bytes, because those bytes are not present. Elasticsearch is eligible only for Firstmate's documented re-pin path after real-document classification. `LANE-ADOPTION.md` gives each lane's next exact step; this branch does not import their unmerged connector files under Firstmate inbox 004.

## Inline GSD lifecycle

- Resolved and applied inline due to the task's no-role-spawning contract: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `scripts/gsd prompt discuss-phase 4347`, `plan-phase 4347 --tdd`, `execute-phase 4347`, `verify-work 4347`, and `code-review 4347`.
- This artifact and `TDD-LEDGER.md` record the red/green/refactor evidence. No GSD role was spawned.
