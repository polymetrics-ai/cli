Closes #3979

## Summary

- Add a PostgreSQL-private exported-snapshot bootstrap coordinator that hands a bounded typed snapshot to pgoutput-v2 at one logical-slot LSN.
- Bind initial LSN, source system/timeline, publication, relation, and schema fingerprint into the durable bootstrap checkpoint; reject any resume drift with typed rebootstrap.
- Add live PostgreSQL 14+ proofs for concurrent snapshot mutation, explicit delete handling, initial checkpoint ordering, slot state/LSN acknowledgement, receipts, and failure/rebootstrap recovery.

## Evidence

- **Red:** `TestPostgresBootstrapCheckpointBindsBarrierAndSchemaFingerprint` initially failed because no bootstrap checkpoint binding existed; the compile failure is in `traces/red-bootstrap-checkpoint.txt`.
- **Green:** focused package tests and the direct Docker/Colima `databaseintegration` command pass. The live proof blocks the snapshot receiver, commits update/delete/insert while it is blocked, then proves the combined keyed result equals the live source exactly.
- **GSD:** manual inline fallback recorded because issue `3979` is not a roadmap phase (`gsd-sdk query init.phase-op 3979` reported `phase_found=false`). Resolved prompts: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`. Required Go skills: `golang-how-to`, design patterns, structs/interfaces, error handling, security, safety, testing, context, concurrency, and database.

## Verification

- `go test -timeout 20m ./internal/connectors/native/postgres/...`
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- `go vet ./internal/connectors/native/postgres/...`
- `go build ./cmd/pm`
- `go run ./cmd/agentcontractgen check`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`

## Scope and safety

- One target connector: native PostgreSQL. No generic source-to-target bridge, raw SQL surface, target DML, mapping/workset change, CLI surface, dependency, or excluded #4125/#4136/#4090 change.
- Snapshot records are receiver-owned copies and initial state advances only after the receiver's connection-owned WAL/Parquet durability boundary returns successfully.
- Follow-ups remain #3983, #4094, and #4095.

## Review

- Inline/manual code review completed and recorded in `REVIEW.md`; the delivery brief makes CI the required gate, so no automated reviewer was requested.
