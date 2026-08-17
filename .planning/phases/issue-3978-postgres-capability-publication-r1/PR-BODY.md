Refs #3978

## Summary

- Publish PostgreSQL `write=true`, `cdc=true`, and `query=false` only where the generated certification matrix has matching accepted live proof.
- Treat `write` as the existing definition-owned, warehouse-mediated managed-target transport. The legacy direct `Connector.Write` remains unsupported and cannot bypass plan, preview, approval, or receipt boundaries.
- Add proof import and matrix projection for the existing PostgreSQL 14+ pgoutput route: independently read a warehouse row, require bounded durable staging and a connection-owned receipt, then observe source acknowledgement at or after the inserted transaction LSN.
- Regenerate the PostgreSQL certification shard and connector/website projections. The global matrix remains honestly incomplete because unrelated applicable cells still lack required fixture/live proof.

## #3978 stale-text reconciliation

1. `incremental_dedupe_history` is executable: the existing PostgreSQL transport evidence covers all six declared modes, including history. This PR does not disable it.
2. Docker is used through the explicit Colima socket. `native/dbtest` accepts it; neither Docker nor Colima was restarted or reconfigured.
3. The issue's previous typed/non-executable CDC wording is stale. Current PostgreSQL change capture is executable only as database source -> connection-owned warehouse. There is still no CDC-to-API or destination `change_capture` route; those cells are concrete N/A/deferred results, never passes.

## Evidence and publication boundary

- Twelve inherited, redacted PostgreSQL transport records provide independent source/target read-back for the six warehouse-mediated modes in both directions. `write` is promoted only when all six exact destination-mode records exist.
- Fresh PostgreSQL 16 built-binary CDC proof emitted two redacted records: `capability:cdc` and `change_capture/database_read_into_warehouse`. It requires the warehouse row and checkpoint, a durable receipt with no pending stage manifest, then acknowledgement at/after the committed LSN.
- `query` remains false. `capability_complete=false` remains correct because it summarizes other applicable cells without complete fixture/live evidence, not a shortfall in the published `write`/`cdc` claims.

## Cases exercised

- Happy: all six managed target modes independently read back; a fresh `pm` captures PostgreSQL ID 901 into Parquet, persists receipt/checkpoint, then acknowledges its LSN.
- Bad: an acknowledgement-before-receipt report is rejected; removing accepted destination evidence keeps `write` unimplemented; direct `Connector.Write` remains unsupported.
- Edge: a schema-valid scratch `sslmode=bananas` profile makes the real certification binary fail after compilation, then green after exact restoration. A startup replication-slot LSN is not accepted as the transaction acknowledgement.

## Verification

```text
go test -timeout 20m ./cmd/connectorgen -count=1                                      PASS
go test -timeout 20m ./internal/connectors/native/postgres -count=1                  PASS
go test -timeout 20m ./internal/connectors/engine -count=1                           PASS
go test -timeout 20m ./internal/connectors/certify -count=1                          PASS
go test -timeout 20m ./internal/app -count=1                                         PASS
go test -timeout 20m ./internal/cli -count=1                                         PASS
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
  go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres \
  -run '^TestPostgresCertificationProfileRunsBuiltBinaryLive$' -count=1 -v           PASS
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
  POLYMETRICS_WRITE_POSTGRES_CERTIFICATION_EVIDENCE=1 \
  go test -timeout 20m -tags=databaseintegration ./internal/cli \
  -run '^TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse$' -count=1 -v        PASS
go vet ./...                                                                           PASS
go build -o pm ./cmd/pm                                                                PASS
go run ./cmd/connectorgen certification-matrix --check                                PASS
./pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs PASS
make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check \
  connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check \
  connectorgen-certification-matrix connector-boundary connector-canon-check \
  release-workflow-check                                                              PASS
pnpm --dir website run lint typecheck test:unit test:scripts test:e2e build           PASS
pnpm --dir website run gen:docs (twice); git diff --exit-code -- website              PASS
```

The mandatory GSD lifecycle was run inline/manual because the repository's single-worker contract prohibits role spawning. The plan, red/green ledger, verification report, and scratch-failure trace are included in this PR.
