# PLAN — PostgreSQL logical-replication CDC (containment)

> Captain ruling, 2026-08-10: do not expose `change_capture` until PostgreSQL
> 14+ protocol-v2 streaming can stage each source transaction privately with a
> hard quota, discard `StreamAbort`, and wait for a durable whole-transaction
> downstream receipt at `StreamCommit` before acknowledging any LSN. The
> current executor therefore fails closed before opening a replication
> connection. Cursor/timestamp reconciliation is explicitly rejected.

## Scope and integration base

The original implementation work targeted a native PostgreSQL logical-replication `ChangefeedExecutor` on
`fm/cli-postgres-cdc-logical-replication-r1`. The branch is rebased onto
`origin/main` after PR #3880 established the non-webhook executor shape and
PR #3882 established the durable database sync contract. This lane changes
only the PostgreSQL native connector, its bundle/docs, the narrow typed
checkpoint extension needed to consume #3882, tests, dependency files, and
this evidence.

`pm` currently has no standalone CDC invocation command. This lane therefore
does not invent a generic SQL/replication CLI surface. Until the committed-
transaction stage exists, the catalogue/inspect projection remains fail-closed
and is the public capability proof.

## Captain execution-proof amendment — 2026-08-10

### Recorded plan gap

There is **no product path for a GitHub-to-PostgreSQL sync today**:
`native/postgres.Connector.Write` returns `ErrUnsupportedOperation`, and the
engine has no database-write executor. A single end-to-end GitHub → PostgreSQL
test would therefore prove neither the GitHub reader nor PostgreSQL CDC. This
is an explicit scope boundary, not a skipped test or a reason to invent a
generic SQL write surface.

### Required proof, deliberately split in two

1. **GitHub read/rate-limit half.** Build and invoke the real `pm` binary
   against read-only public `rails/rails` issues or pull requests, requesting a
   high-volume stream. Record the observed request count and wall-clock rate,
   whether the connector limiter imposed a wait, and whether GitHub returned
   403 or 429. Do not configure, print, store, or depend on a token, and do not
   perform a GitHub write.
2. **PostgreSQL CDC half.** Start an isolated **PostgreSQL 14+** container using
   Docker through Colima (never Podman for this proof), with `wal_level=logical`.
   Load the first half's non-secret public record data directly through an
   explicit local test loader/`psql` path because no product write executor
   exists. The live conformance test must prove ordered inserts, updates, and
   deletes; byte-exact non-ASCII text; a mid-stream restart/resume from the
   durable source-bound checkpoint; and connector-owned slot teardown leaving
   no pinned replication slot. It must run under `POLYMETRICS_INTEGRATION=1`;
   no test may remain disabled merely because CDC was formerly planned.

The staged executor remains subject to the earlier captain ruling: PostgreSQL
14+ only, `proto_version=2` with streaming enabled, a bounded
crash-recoverable private source-transaction stage, `StreamAbort` discard, and
source progress only after the downstream port durably accepts the **whole**
transaction at `StreamCommit`. A hard quota raises
`TransactionStageLimitExceeded` without acknowledgement. Slot-health
observability and an explicit connector-owned teardown/rebootstrap procedure
ship with the admission change. No cursor or timestamp reconciliation may
substitute for change capture. `capabilities.cdc` remains false until every
live proof in this amendment and the staged-transaction contract passes.

### TDD and delivery checkpoints for this amendment

1. **Red — executable admission:** add a regression showing the planned reader
   cannot be promoted until it requires PostgreSQL 14+, protocol v2 streaming,
   a bounded stage, and a durable whole-transaction receipt.
2. **Red — real system proof:** make the PostgreSQL integration suite fail when
   `POLYMETRICS_INTEGRATION=1` if its Docker/Colima server cannot start, if the
   disabled historical conformance test is not enabled, or if a post-teardown
   slot remains. Add an explicit public-read runner that reports only aggregate
   request/rate/status statistics.
3. **Green — implementation and proof:** implement the smallest native
   PostgreSQL-only staged executor and teardown/rebootstrap procedure that
   satisfies those red tests, run both halves against real services, and record
   exact aggregate measurements in `VERIFICATION.md` and the PR body.
4. **Promotion checkpoint:** regenerate every derivable connector artifact;
   flip `changefeed.json` and `capabilities.cdc` only after the green live
   proof, then rerun capability, help/manual/website parity, and conformance
   checks. If any proof is unavailable or fails, leave the capability false and
   report the named blocker.

### Execution-blocking foundation gaps confirmed — 2026-08-10

The first proof half is independently executable and has been run. The second
half is **not safely implementable inside this connector lane yet**, for two
concrete reasons:

1. `connectors.CDCReadRequest` exposes only a per-record
   `emit func(CDCEvent) error` callback and an envelope-only
   `DurableChangefeedCheckpointCommitter`. It has no
   committed-transaction/receipt port that can atomically accept one sealed
   source transaction at `StreamCommit`. Adding the required
   `CommittedLogicalTransactionSink` (or equivalent) changes shared
   `internal/connectors` runtime contract, which the issue-first connector
   contract requires be split into a foundation issue/PR rather than absorbed
   here.
2. At this plan's 2026-08-10 baseline, the merged `native/dbtest` harness was
   explicitly Podman-only. The separate #4083 foundation now supplies the
   current explicit Docker-or-Podman direct-local-Unix contract, including the
   Docker/Colima capacity proof; its [maintainer guide](../../../internal/connectors/native/dbtest/README.md)
   owns the current details. A hand-rolled Docker command path would still
   bypass the harness's resource-ownership and cleanup contract.

With `POLYMETRICS_INTEGRATION=1`, the retained historical conformance test
still exits at its unconditional planned-CDC skip before it can contact a
server. The focused fail-closed regression passes. These are evidence of
containment, not CDC conformance. Do not remove the skip, start a v1 reader,
or flip the capability merely to make a live-test command appear green.

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
`golang-documentation`, and `golang-lint`. The CLI/help/docs/website parity
reference was also read: no command, help, manual, or website page is added,
while the existing connector docs and capability tests are updated.

## Historical design decisions

1. `pglogrepl` is pinned only after the recorded conditional-approval evidence
   in `PR-BODY.md`. It is used solely by `native/postgres` for PostgreSQL
   replication protocol commands; no other new module is admitted.
2. A replication connection sets `replication=database`, calls
   `IDENTIFY_SYSTEM`, and uses the server system identifier plus database and
   requested stream as the `synccontract.SourceIdentity`. The server timeline
   plus validated publication name is the source generation. A previous
   `CheckpointEnvelope` is validated with `ValidateResume` before replication
   begins; a scalar `State["lsn"]` is not accepted as a parallel checkpoint
   store.
3. The generated, source-bound slot name is `pm_cdc_<hash>` and cannot be
   caller-selected. It is stable across restarts for the same server/database/
   stream, creating once and reusing only a compatible inactive `pgoutput`
   logical slot. The explicit native teardown method drops only that derived
   slot. Tests always defer teardown and assert the slot is gone: a routine
   that leaves a WAL-retaining slot is not conforming.
4. The parked protocol path has the caller declare a pre-existing PostgreSQL
   publication by a validated identifier and exposes no generic SQL or DDL
   surface. Its `pgoutput` v1 decoder can map Relation/Insert/Update/Delete
   frames into the existing `CDCEvent` record shape, but `ReadCDC` deliberately
   cannot reach that path: it has no streamed whole-transaction stage.
5. The parked path creates a `synccontract.CheckpointEnvelope` candidate at
   commit and hands it to the existing durable checkpoint committer only after
   its records were accepted. That is not sufficient admission evidence for
   v2 streamed transactions: a future stage must wait for a sealed source
   transaction and its whole-transaction downstream receipt before using the
   same committer or acknowledging a source LSN.
6. The historical descriptor declared `native/postgres_logical_replication`,
   `lsn`, `downstream_ack`, `resnapshot_required`, source ordering, at-least-
   once duplicate semantics, and tombstones. The current descriptor instead
   declares `planned`, so the existing `HasImplementedChangefeed` gate remains
   unchanged and PostgreSQL cannot be discovered as CDC-capable.

## Historical TDD sequence

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
6. **Post-rebase hardening:** retain only a source-bound native checkpoint,
   preserve its original slot barrier across restart, refuse an existing slot
   without that checkpoint, validate publication membership before slot
   creation, filter multi-table publications to the requested relation, and
   map full-table truncation as an explicit empty-record CDC event. Prove this
   against a real PostgreSQL source as well as focused decoder tests.

## Exclusions

- MySQL, MariaDB, SQL Server, Oracle, generic raw SQL, HTTP-write, shell, and
  redaction work.
- Any weakening of `HasImplementedChangefeed`, command preflight, or the
  requirement for durable acknowledgement before source LSN acknowledgement.
- Automatic publication creation/deletion. Operators own the declared
  publication; the connector owns only its uniquely derived replication slot.
- Any at-commit transaction burst, acknowledgement on local stage durability,
  or automatic cursor/timestamp recovery after a stage-limit failure.
