# Plan — #3979 PostgreSQL gap-free snapshot-to-changefeed bootstrap

## Delivery and lifecycle

This is the explicit inline/manual-GSD fallback described in `CONTEXT.md`: the active roadmap has no phase `3979`, and the canonical contract requires one worker. The required command path was resolved through `scripts/gsd`: `discuss-phase 3979 --auto`, `plan-phase 3979 --tdd --auto`, `execute-phase 3979 --interactive --auto`, `verify-work 3979 --auto`, and `code-review 3979 --auto`. No role was spawned.

Required skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.

## Goal

Bridge PostgreSQL's exact bounded snapshot to executable pgoutput-v2 change capture without skipping a commit made while the snapshot is materialized, without duplicate final rows, and without advancing a slot before the snapshot and downstream receipt are durable.

## TDD slices

1. **Red — bootstrap boundary contract.** Add focused PostgreSQL tests for an exported slot snapshot, a snapshot receiver failure, and the initial durable-barrier ordering. The test must observe that no snapshot receipt/checkpoint/standby acknowledgement happens before the corresponding callback succeeds, and that a retained uncheckpointed slot cannot be silently reused.
2. **Green — bootstrap coordinator.** Add the smallest PostgreSQL-private coordinator/request types. It performs existing CDC preflight/identity/publication checks, creates an exported-snapshot logical slot, imports that snapshot into the existing typed repeatable-read page plan, and supplies bounded pages to the receiver. It validates source, timeline, publication, relation and schema bindings before any source page is delivered.
3. **Green — handover.** After the snapshot receiver has confirmed durable materialization, create the initial envelope at the slot LSN, persist it through the supplied durable committer, acknowledge only that persisted barrier, then start the existing pgoutput-v2 machine from the same LSN. Preserve all existing receipt -> checkpoint -> acknowledgement behaviour for post-barrier transactions.
4. **Red/live — overlapping mutation proof.** Extend the existing PostgreSQL `databaseintegration` harness with a table seeded before barrier, a receiver that deliberately blocks after the barrier, a concurrent multi-row update/delete/insert, and a post-snapshot write. Apply the snapshot and emitted change events to a keyed oracle. Assert exactly the final server rows, explicit delete tombstone, barrier checkpoint, and acknowledged slot LSN. The test fails if the implementation starts its snapshot before the slot barrier or starts changefeed after the wrong LSN.
5. **Green/live — restart and refusal proof.** Exercise receiver/committer failure and retry, source/publication/schema drift guards, and slot cleanup. Every refusal asserts no downstream page/CDC callback, no durable checkpoint, and no LSN acknowledgement beyond the prior durable point.

## Guardrails

- Production changes are restricted to `internal/connectors/native/postgres/**`; evidence lives only in this issue directory.
- No generic source/destination bridge, target DML, mapping/workset change, schema migration, CLI/docs change, new dependency, raw SQL caller surface, or connector metadata promotion.
- PostgreSQL-generated snapshot tokens are treated as data and quoted only in the private imported-snapshot transaction. All operator inputs retain existing identifier validation and parameterization.
- Slot creation precedes snapshot observation; an old slot without the durable bootstrap checkpoint is rebootstrap-required, never assumed to be a valid boundary.
- Page and stage limits remain the existing finite typed-catalog and committed-transaction-stage limits; the coordinator must not collect the snapshot into one slice.

## Commit checkpoints

1. GSD planning evidence.
2. Preserved focused red test output.
3. Green coordinator plus focused package tests.
4. Green live PostgreSQL/dbtest proof and review-fix checkpoint.
