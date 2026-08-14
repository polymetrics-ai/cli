Refs #3983

## Summary

- derive a workset-bound, managed-target `incremental_upsert` plan using the shared `MappingContractV1`, `DeliveryReceiptV1`, and `DatabaseWriteExecutor`
- require a one-shot exact approval, deliver only delta rows plus explicit mapped tombstones, and make unknown commit replay-required
- persist and verify a per-destination candidate baseline only after the durable target receipt
- live-prove PostgreSQL insert/update, physical-absence retention, explicit tombstone deletion, ledger receipt, and baseline promotion

## Evidence

| Criterion | Evidence |
| --- | --- |
| Physical absence never deletes | Live PostgreSQL query retains `retain/9` after it is omitted from the next source projection. |
| Explicit tombstone deletes | Live PostgreSQL query removes only `retain/9` after the sealed tombstone. |
| Stale approval has zero writes | Fake session, ledger, and baseline counters all remain zero. |
| Failure/unknown retain baseline | Deterministic fakes retain the prior baseline on batch failure, receipt/ledger failure, and unknown commit. |
| Destination isolation | Concurrent distinct owner/connection destinations re-read only their own candidate baseline. |

## Verification

- `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/warehouse/... -count=1`
- `go test -race -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test(ChangeDelivery|FileChangeDelivery|PostgresManagedTargetWorksetDelivery)' -count=1`
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- `go vet ./...`, `go build ./cmd/pm`, scoped lint, repository non-suite gates, and `git diff --check`

## Delivery record

- GSD lifecycle: manual inline `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`; generated prompts resolved and evidence is in `.planning/phases/issue-3983-postgres-workset-delivery-r1/`.
- Fallback reason: canonical single-worker / Firstmate direct-PR delivery forbids spawning lifecycle roles.
- Required skills: `golang-how-to`, design patterns, structs/interfaces, error handling, security, safety, testing, database, context, concurrency, and linting.
- CLI help/manual/website parity: not applicable; no command, flags, help topic, generated connector bundle, or website surface changed.
