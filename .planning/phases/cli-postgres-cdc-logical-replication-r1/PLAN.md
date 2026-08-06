# PLAN — PostgreSQL logical-replication CDC

## Scope and integration base

Implement the native PostgreSQL logical-replication `ChangefeedExecutor` on
`fm/cli-postgres-cdc-logical-replication-r1`. The branch is rebased onto
`origin/main` after PR #3880 established the non-webhook executor shape and
PR #3882 established the durable database sync contract. This lane changes
only the PostgreSQL native connector, its bundle/docs, the narrow typed
checkpoint extension needed to consume #3882, tests, dependency files, and
this evidence.

`pm` currently has no standalone CDC invocation command. This lane therefore
does not invent a generic SQL/replication CLI surface. It makes the registered
`postgres` connector genuinely executable through `ReadCDC`; the existing
fail-closed catalogue/inspect projection is the public capability proof.

## GSD and skills

`scripts/gsd doctor`, all required `scripts/gsd sources` resolutions,
`go run ./cmd/agentcontractgen check`, and generated prompts for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` passed/are recorded. This is an issue/branch lane rather than a
roadmap-numbered phase, and the canonical single-worker contract forbids role
spawning, so the generated lifecycle is followed inline through these durable
artifacts.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-concurrency`, `golang-database`, `golang-dependency-management`, and
`golang-documentation`. The CLI/help/docs/website parity reference was also
read: no command, help, manual, or website page is added, while the existing
connector docs and capability tests are updated.

## Design decisions

1. `pglogrepl` is pinned only after the recorded conditional-approval evidence
   in `PR-BODY.md`. It is used solely by `native/postgres` for PostgreSQL
   replication protocol commands; no other new module is admitted.
2. A replication connection sets `replication=database`, calls
   `IDENTIFY_SYSTEM`, and uses the server system identifier plus database and
   requested stream as the `synccontract.SourceIdentity`. The server timeline
   is the source generation. A previous `CheckpointEnvelope` is validated with
   `ValidateResume` before replication begins; a scalar `State["lsn"]` is not
   accepted as a parallel checkpoint store.
3. The generated, source-bound slot name is `pm_cdc_<hash>` and cannot be
   caller-selected. It is stable across restarts for the same server/database/
   stream, creating once and reusing only a compatible inactive `pgoutput`
   logical slot. The explicit native teardown method drops only that derived
   slot. Tests always defer teardown and assert the slot is gone: a routine
   that leaves a WAL-retaining slot is not conforming.
4. The caller declares the pre-existing PostgreSQL publication by a validated
   identifier. The connector exposes no generic SQL or DDL surface. It starts
   `pgoutput` with protocol version 1, decodes Relation/Insert/Update/Delete
   frames, and emits the existing `CDCEvent` record shape.
5. A commit frame becomes one `synccontract.CheckpointEnvelope` candidate. It
   is handed to a typed durable checkpoint committer only after all transaction
   records were accepted, and the server receives a standby status update only
   after that durable commit succeeds. The committer is the adapter that uses
   `CommitAfterDownstreamAcknowledgement`; its full envelope replaces, rather
   than serializes alongside, the old state map.
6. The descriptor declares `native/postgres_logical_replication`, `lsn`,
   `downstream_ack`, `resnapshot_required`, source ordering, at-least-once
   duplicate semantics, and tombstones. `Connector` implements the matching
   executor descriptor, so the existing `HasImplementedChangefeed` gate is
   unchanged and becomes true only for this working implementation.

## TDD sequence

1. **Red — admission and durable state:** replace the stub expectations with
   a `ChangefeedExecutor`/catalogue truth test; add source-identity mismatch,
   scalar-state rejection, and durable-commit-before-status tests.
2. **Red — lifecycle and protocol:** add seam tests for deterministic derived
   slot naming, compatible reuse, incompatible/active-slot refusal, cleanup,
   transaction commit boundaries, and no checkpoint on emit/commit failure.
3. **Green — native implementation:** add the typed checkpoint port, stable
   source identity/LSN envelope construction, replication connection, slot
   lifecycle, `pgoutput` decoding, acknowledgement ordering, and teardown.
4. **Green — live conformance:** an integration test gated by
   `POLYMETRICS_INTEGRATION=1` connects to a real PostgreSQL with
   `wal_level=logical`, creates a fixture table/publication, reads insert,
   update, and delete changes, restarts from the committed envelope without
   loss or duplicate, then proves the derived slot is absent after teardown.
   It reads its credential only from environment and never prints it.
5. **Documentation and proof:** update the PostgreSQL bundle metadata,
   changefeed evidence, and docs; measure the post-dependency binary; run
   focused tests plus all non-suite gates; record review and parity evidence.

## Exclusions

- MySQL, MariaDB, SQL Server, Oracle, generic raw SQL, HTTP-write, shell, and
  redaction work.
- Any weakening of `HasImplementedChangefeed`, command preflight, or the
  requirement for durable acknowledgement before source LSN acknowledgement.
- Automatic publication creation/deletion. Operators own the declared
  publication; the connector owns only its uniquely derived replication slot.
