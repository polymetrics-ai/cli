# Verification — Issue #3978: final PostgreSQL certification and publication

## Scope and reconciliation

This is a certification/publication change. It does not add a generic SQL
writer, a direct source-to-target hop, an API CDC sink, a broker, MCP, or UI
behavior.

- The issue's request to reject `incremental_dedupe_history` is stale. The
  current PostgreSQL definition declares all six managed-target modes and the
  current committed evidence has a real-binary record for each; the matrix now
  retains that fact.
- The issue's Podman-only wording is stale. All live checks used the explicitly
  configured shared Colima Docker socket; Docker/Colima were neither restarted
  nor reconfigured.
- The pre-task `write=false, cdc=true` publication was stale relative to merged
  #4195/#4199 transport behavior. `write=true` now means only the declared
  warehouse-mediated managed destination. It never makes direct
  `Connector.Write` available. `query=false` remains exact.
- PostgreSQL CDC remains a source-to-connection-owned-warehouse route only.
  The generated matrix records API primitives and destination `change_capture`
  as non-pass cells with concrete reasons. No CDC-to-API behavior is implied.

The four programme transfer directions are the previously merged #3987/#4195
evidence set (API→API, API→database, database→API, database→database). This
issue's fresh live work re-certifies the PostgreSQL definition-owned six-mode
database transport and source-only CDC publication; it does not re-open the
GitHub/external-binary lanes that the task excludes.

## TDD and failure controls

Red observations:

1. `TestCertificationMatrixPromotesPostgresManagedWriteOnlyWithDeclaredLiveTransportProof`
   failed before publication because PostgreSQL was still declared `write=false`.
2. `TestCertificationMatrixPromotesPostgresChangeCaptureOnlyWithReceiptBackedLiveProof`
   failed before the fresh CDC proof because the accepted CDC capability record
   was absent.
3. `TestCertificationChangeCaptureRequiresImplementedChangefeedContract`
   failed against a source whose bundle changefeed had been removed. The green
   projection now refuses to declare that mode without an implemented database
   changefeed contract.
4. The scratch `sslmode=bananas` profile compiled through the definition
   validator but the real built PostgreSQL certification binary exited 2. The
   exact run and restoration are in `traces/red-runtime-profile.txt`.

Green controls:

- `TestCertificationMatrixKeepsPostgresManagedWriteUnimplementedWithoutAllDeclaredLiveProof`
  proves a metadata declaration plus an empty evidence set is insufficient.
- `TestPostgresPublishedWriteAndCDCMatchLiveCertification` requires every true
  PostgreSQL public flag to have a declared, implemented, accepted live matrix
  cell. `write` points at six exact destination-mode records; `cdc` points at
  the receipt-backed record.
- `TestWriteUnsupported` remains green: a direct `Connector.Write` cannot
  bypass the managed transport's plan/preview/approval/receipt boundary.

## Live PostgreSQL evidence

```text
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres \
  -run '^TestPostgresCertificationProfileRunsBuiltBinaryLive$' -count=1 -v
PASS

POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
POLYMETRICS_WRITE_POSTGRES_CERTIFICATION_EVIDENCE=1 \
go test -timeout 20m -tags=databaseintegration ./internal/cli \
  -run '^TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse$' -count=1 -v
PASS
```

The certification profile builds `pm`, executes the declared PostgreSQL
transport, verifies all six modes with independent target SQL read-back, and
requires positive read/load/checkpoint counts. The CDC binary test builds a
fresh `pm`, starts a PostgreSQL 16 pgoutput source, commits ID 901, reads that
ID back from the connection-owned Parquet warehouse, requires a durable receipt
with no pending transaction manifest and a complete checkpoint, then observes
`confirmed_flush_lsn` at or after the post-insert transaction LSN. It emitted
only these redacted proof records:

- `internal/connectors/certifications/evidence/postgres_cdc_r1-capability-cdc.json`
- `internal/connectors/certifications/evidence/postgres_cdc_r1-change_capture-database_read_into_warehouse.json`

The generated PostgreSQL matrix therefore has `write=true` with six exact live
destination records and `cdc=true` with one receipt-backed capability record
plus one `change_capture/database_read_into_warehouse` record. `query` is
false. Its aggregate `capability_complete` remains false because unrelated
applicable capability cells lack fixture/live evidence; that aggregate is not a
substitute for the individual published capability evidence.

## Generated artifacts and CLI parity

```text
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
PASS

pnpm --dir website run gen:website-data
PASS
pnpm --dir website run gen:docs
PASS
pnpm --dir website run gen:docs
PASS
git diff --exit-code -- website
PASS (no second-pass working-tree diff)

./pm help connectors
./pm connectors
./pm connectors certify --help
PASS (all contextual help succeeds)

./pm connectors inspect postgres --json
PASS: connector.capabilities.write=true, cdc=true, query=false
```

## Local verification

```text
go test -timeout 20m ./cmd/connectorgen -count=1                                      PASS
go test -timeout 20m ./internal/connectors/native/postgres -count=1                  PASS
go test -timeout 20m ./internal/connectors/engine -count=1                           PASS
go test -timeout 20m ./internal/connectors/certify -count=1                          PASS
go test -timeout 20m ./internal/app -count=1                                         PASS
go test -timeout 20m ./internal/cli -count=1                                         PASS
go vet ./...                                                                           PASS
go build -o pm ./cmd/pm                                                                PASS
make tidy-check                                                                        PASS
make lint                                                                              PASS (0 issues)
make docs-check-no-build                                                               PASS
make smoke-no-build                                                                    PASS
make agent-contract-check                                                              PASS
make connectorgen-validate                                                             PASS (552 bundles, 0 findings)
make connectorgen-surface-sync                                                        PASS (0 drift)
make github-parity-artifacts-check                                                     PASS
make connectorgen-certification-matrix                                                 PASS
make connector-boundary                                                                PASS
make connector-canon-check                                                             PASS
make release-workflow-check                                                            PASS
pnpm --dir website run lint                                                            PASS (13 existing warnings)
pnpm --dir website run typecheck                                                       PASS
pnpm --dir website run test:unit                                                       PASS (80 tests)
pnpm --dir website run test:scripts                                                    PASS (28 tests)
pnpm --dir website run test:e2e                                                        PASS (19 passed, 7 expected skips)
pnpm --dir website run build                                                           PASS
```

## Inline verification and review fallback

The canonical contract prohibits role spawning in this worker. I ran the
GSD adapter health/source checks and generated the `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts,
then performed the required TDD, verification, and standard cross-file review
inline. Review focused on the evidence importer, proof redaction boundary,
matrix projection, source LSN ordering, generated artifacts, and the direct
writer boundary. The global matrix check caught one real GitHub regression
during review; it was fixed by limiting the new projection to declared native
database destinations, then rechecked byte-identical for GitHub. No unresolved
Critical or Warning finding remains.
