# Context — #3979 PostgreSQL gap-free bootstrap

## Task Delivery Header

- Issue: `Closes #3979 — Postgres Parity: add gap-free snapshot-to-changefeed bootstrap`
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 -> main`
- Delivery: PR open against `integration/4015-mvp-flat-r1` with green checks.
- Working branch: `fm/cli-3979-gapfree-bootstrap-r1`
- Task: Make the snapshot-to-changefeed handover gap-free.
- Verification: `go test -timeout 20m ./internal/connectors/native/postgres/...` and the explicit PostgreSQL `dbtest` Docker/Colima run.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Writes before, during, and after bootstrap form one correct final relation without a boundary gap or duplicate key. | live | A PostgreSQL 14+ `dbtest` fixture blocks the snapshot receiver after the exported slot barrier, mutates rows concurrently, then applies the snapshot and pgoutput events by primary key and asserts the final live relation exactly matches the combined result. |
| Transactions are atomic and deletes are explicit tombstones. | live + fake | The live fixture commits a multi-row update/delete transaction and observes no partial callback before commit; a deterministic callback fake is necessary to make the receipt/checkpoint/acknowledgement ordering observable. |
| Failure before a durable bootstrap checkpoint cannot advance the slot or silently resume. | fake | A receiver/committer failure records zero durable checkpoint and acknowledgement calls; retry against the retained uncheckpointed slot returns typed `RequireRebootstrap` rather than recreating or cursor-falling back. |
| Source identity, publication, schema fingerprint, and initial LSN are bound to the bootstrap state. | live + fake | Live fixture compares the slot barrier and source identity against PostgreSQL; focused tests vary source/schema/publication inputs and assert the bootstrap refuses before snapshot callback or replication start. |
| Explicit integration opt-in runs a PostgreSQL dbtest harness rather than skipping. | live | `POLYMETRICS_DATABASE_INTEGRATION=1` with the supplied direct Docker socket starts the harness, mutates rows concurrently, and asserts rows/checkpoint/LSN/slot state rather than process exit. |

## Locked decisions

1. The sole production connector target is native PostgreSQL. Reuse the landed bounded snapshot reader, logical-slot lifecycle, pgoutput-v2 transaction machine, and committed transaction stage; do not alter shared connector, mapping, workset, destination, or CLI contracts.
2. Bootstrap creates the connector-owned logical slot with an **exported** PostgreSQL snapshot. The exported snapshot is imported into one read-only repeatable-read transaction, so catalog fingerprint and paged rows are observed at the same slot-consistent boundary.
3. Snapshot pages are delivered through a narrow PostgreSQL-owned receiver that returns only after its connection-owned WAL/Parquet materialization is durable. No direct connector-to-destination call, raw SQL API, generic receiver interface, or unbounded in-memory collection is introduced.
4. After the snapshot receiver is durable, persist the initial CDC checkpoint at the slot barrier, then acknowledge that barrier and begin pgoutput-v2 from it. A committed transaction after the barrier continues through the existing receipt -> checkpoint -> acknowledgement path.
5. A pre-existing slot without a durable bootstrap checkpoint remains a typed rebootstrap requirement. Slot loss, retention loss, source system/timeline, publication, or relation/schema fingerprint drift fail closed; they never recreate a cursor or silently adopt a new slot.
6. The integration test must force a true overlap: it blocks snapshot delivery after the barrier, commits writes in that window, and verifies final keyed state after the changefeed has applied. Sequential quiet-database tests are not accepted evidence.

## GSD lifecycle fallback

`scripts/gsd doctor`, every required `scripts/gsd sources` query, all five generated lifecycle prompts, and `go run ./cmd/agentcontractgen check` passed. `gsd-sdk query init.phase-op 3979` reports `phase_found: false`, because the active numbered roadmap has no #3979 phase. The canonical single-worker contract prohibits role spawning. This issue-scoped context, discussion log, plan, TDD ledger, verification record, and review record are the explicit inline/manual-GSD fallback.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`. The PostgreSQL runtime/dbtest integration reference was read. No CLI surface changes are planned, so help/manual/website parity is not applicable.

## Canonical references

- `internal/connectors/native/postgres/transport_source.go` — bounded repeatable-read snapshot and typed page plan.
- `internal/connectors/native/postgres/cdc.go` — executable pgoutput-v2 CDC entry point and checkpoint validation.
- `internal/connectors/native/postgres/cdc_lifecycle.go` — PostgreSQL preflight, source identity, slot lifecycle, and typed rebootstrap failures.
- `internal/connectors/native/postgres/cdc_v2.go` — staged transaction receipt/checkpoint/acknowledgement ordering.
- `internal/connectors/native/postgres/cdc_integration_test.go` — PostgreSQL 14+ dbtest harness and observable slot/LSN assertions.
- `internal/connectors/database/parquet_delivery_workset.go` — immutable downstream workset foundation, intentionally not altered by this source-only integration.
- `internal/connectors/database/mapping_contract.go` — landed shared mapping contract, intentionally not altered.
- `internal/connectors/native/dbtest/README.md` — mandatory explicit runtime and cleanup harness contract.
- `docs/architecture/connector-architecture-v2-design.md` — native connector boundary.

## Deferred / out of scope

- #3983 target delivery, #4094 history target, #4095 CDC-to-target binding, and the expressly excluded #4125, #4136, and #4090 work.
- New dependencies, generic CDC/bootstrap interfaces, connector metadata changes, CLI commands, documentation surfaces, or generic SQL/write tools.
