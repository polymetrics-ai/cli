# PLAN — Issue 3810: shared database sync contract

## GSD setup and fallback

- `scripts/gsd doctor` passed in this worktree.
- Resolved command sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed.
- #3810 is an issue foundation, not a numbered phase in `.planning/ROADMAP.md`, so the generated
  GSD phase workflows cannot initialize it. Manual inline execution is used as permitted by the
  project contract and the task's single-worker constraint. This artifact is the discussion/plan
  output; the TDD ledger and verification checklist remain the execution record.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-database`.

## Goal

Replace the app's scalar stream cursor with a shared, versioned sync-state contract and make every
new database sync claim admit execution only through a named native executor plus immutable
conformance evidence.

## Allowed implementation scope

- `internal/synccontract/**` — sole shared owner of modes, state, recovery, acknowledgement,
  tombstone/history, native-contract admission, and embedded reusable fixtures.
- `internal/app/sync_modes.go`, `internal/app/types.go`, `internal/app/app.go`,
  `internal/app/local_warehouse.go` and focused tests — compatibility adapters and state-envelope
  persistence only.
- `.planning/phases/cli-found-database-sync-contract-r1/**` — planning, red/green, and verification
  evidence.

## Explicit exclusions

- No database engine, bundle declaration, changefeed descriptor/schema, connectorgen, CLI/docs/
  website surface, generic transport, credential, raw SQL, raw HTTP, shell, or redaction work.
- No `commandrunner` edits. Its parallel ownership split remains untouched.
- Do not reinterpret or implement the local warehouse's existing dedupe materializer; the shared
  history contract merely makes a physical delete non-conforming for future history consumers.

## Design

### Shared package

`internal/synccontract` is dependency-free and therefore importable by both `app` and future
connector/transport lanes without an import cycle. It supplies:

1. `Mode` with exactly seven contract names, plus validation and an explicit native-executability
   admission function.
2. `CheckpointEnvelope` and nested source, barrier, position, partition, and dedupe structures.
   Every opaque token is a byte slice and is defensively copied.
3. `RebootstrapRequiredError` with a closed outcome code; validation returns it without mutation.
4. `CommitAfterDownstreamAcknowledgement`, which cannot invoke a state committer until an explicit
   durable acknowledgement has been supplied and stamps `committed_at` separately from
   `observed_at`.
5. `Tombstone` and history-window close mutation rules.
6. `NativeCommandContract` / `NativeSyncExecutor` matching. Its schema has no REST surface or
   generic query payload; only named protocol/executor references and embedded-fixture evidence.
7. An embedded versioned fixture corpus returned as defensive copies for all consumers.

### App compatibility seam

`StreamState.Cursor` is removed. JSON decoding recognizes an old `cursor` key only to preserve its
opaque bytes in a version-zero legacy envelope. The next attempted resume is a typed explicit
rebootstrap requirement, not a cleared state or full scan. New legacy-app runs create version-one
envelopes through the shared contract and write their state only after their existing destination
write/atomic-replace acknowledgement path completes.

The old five names remain an internal compatibility profile. `ParseSyncMode` recognizes the seven
closed names without creating an execution claim; any contract mode without matching native
executor/evidence fails before source read. This prevents the foundation itself from advertising a
capability it cannot execute.

## TDD task sequence

1. **Red:** add contract tests for every required mode/state/recovery/ack/delete/native guarantee
   and app tests for scalar-state migration and no state advance after failed downstream work.
2. **Green:** implement the dependency-free contract package and immutable fixture corpus.
3. **Green:** adapt app state persistence, legacy scalar decoding, and legacy ETL checkpoint
   construction/commit ordering to the contract.
4. **Regression:** retain legacy behavior tests with their expectations changed only from scalar
   field inspection to envelope inspection.
5. **Verification:** focused package tests plus app/connector tests, vet/build, and every
   non-suite `make verify` gate required by `AGENTS.md`; run code review after green.

## Evidence strategy

- Red command output is retained under `traces/red-run.txt` before production code exists.
- The final test corpus exercises opaque byte round trips (including non-UTF-8 data), noncollapsed
  partition state, all typed recovery outcomes, acknowledgement order, history delete close, and
  native execution admission.
- No fixture performs a provider call; conformance data is static and versioned. The current
  implementation registers no native executor, so all seven contract modes correctly remain
  non-executable until a later lane supplies both runtime and evidence.
