# Context — #3977 pgoutput v2 CDC

## Task Delivery Header

- Issue: Refs #3977 — Postgres Parity: make pgoutput v2 change capture executable.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: A pull request open against `integration/4015-mvp-flat-r1`, with the requested local checks and CI green.
- Working branch: `fm/cli-3977-pgoutput-v2-cdc-r1`.
- Task: Make the native PostgreSQL reader execute PostgreSQL 14+ `pgoutput` protocol v2 streamed transactions with bounded durable staging, receipt-before-checkpoint-and-LSN acknowledgement, safe restart/teardown, and live container proof.
- Verification: `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/...`; the explicit Docker/Colima `databaseintegration` PostgreSQL test; connector validation and surface checks; CI.

## Locked decisions

1. PostgreSQL is the only connector in scope. Reuse the landed `database.CommittedTransactionStage`; do not alter shared database or generic connector contracts.
2. Use `pgoutput` protocol version 2 with `streaming 'on'`; PostgreSQL below 14 must fail before replication begins.
3. Relation/type/origin metadata is decoded and retained locally. DML/truncate frames are serialized into bounded private stage chunks. `StreamAbort` discards its transaction and causes no event, checkpoint, or standby-status acknowledgement.
4. `StreamCommit` is the release boundary for streamed transactions. The receiver consumes all ordered chunks, calls the existing event callback, returns a durable downstream receipt, the stage persists its immutable receipt, then the connector persists the checkpoint and sends the commit LSN acknowledgement.
5. Stage material is rooted below the caller's project directory, with source-derived safe names and explicit finite limits. A stage transaction identity includes its source, PostgreSQL XID, and stable WAL position so an old receipt cannot collide after XID reuse. A missing project directory is refused rather than silently falling back to an ephemeral spool.
6. The existing source identity, publication membership, replication-slot, retention/rebootstrap, restart, and teardown safeguards remain in the executable path.
7. This repo's numbered GSD roadmap does not contain issue `3977`, so `gsd-sdk query init.phase-op 3977` reports `phase_found: false`. The lifecycle commands were resolved and prompted, but their normal artifacts cannot be generated. This issue directory is the explicit inline/manual-GSD fallback; no roles were spawned because the canonical contract requires one worker.

## Deferred / out of scope

- Generic CDC receiver APIs, shared runtime contracts, snapshot-to-changefeed bootstrap (#3979), destination writes, and unrelated defects #4125, #4136, and #4090.
- A new CLI command, help topic, manual page, or website page. Existing connector capability metadata is promoted only after the live proof passes.
