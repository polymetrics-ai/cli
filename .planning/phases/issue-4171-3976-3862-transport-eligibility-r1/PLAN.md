# PLAN — transport source eligibility club (#4171, #3976, #3862)

## GSD setup and execution mode

- `scripts/gsd doctor`, lifecycle source resolution, and
  `go run ./cmd/agentcontractgen check` passed in the isolated task worktree.
- `discuss-phase --auto` and `plan-phase --tdd --auto` are executed inline. The task is an
  issue-club phase outside the numbered roadmap, the user requires autonomous single-worker
  delivery, and the canonical contract forbids spawning lifecycle roles. The generated official
  prompts, TDD ledger, verification checklist, and review gate remain authoritative.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-database`.
- The runtime/PostgreSQL integration reference and connector-v2 migration canon were read before
  planning.

## Goal

Close three production routing gaps without changing the registry's fail-closed model:

1. GitHub explicitly admits each declarative stream the source adapter can execute, including
   `commits`, and pages/batches it through the existing warehouse-mediated transport.
2. PostgreSQL's definition-selected native keyset reader is adapted to the shared resumable
   polling executor and reached by production composition.
3. Cross-family conformance is proven from the shipped entry path, while undeclared streams,
   executors, and evidence remain typed pre-I/O refusals.

## Scope and ownership

- `internal/connectors/defs/github/**`: explicit source eligibility and derived bundle artifacts.
- `internal/app/**`: production composition for the general declarative source and focused
  shipped-entry tests.
- `internal/synctransport/**`: typed stream-ineligibility outcome without changing admission
  order, exact executor lookup, or definition-owned evidence verification.
- `internal/connectors/defs/postgres/**`, `internal/connectors/native/postgres/**`, and the narrow
  shared polling seams in `internal/connectors/engine/**`: implemented polling declaration,
  dynamic native binding, keyset fetch/traversal, and shared executor invocation.
- This phase directory: red/green traces, verification, review, summary, and PR evidence.

No CLI command, flag, or output shape is added. `max_pages` remains a connector runtime-config
key, so CLI help/manual/website parity is not applicable; connector authoring and operator docs are
updated only if the existing generated bundle/docs surface changes.

## Non-goals

- No edits for #4125, #4158, or #4169.
- No capability-bit inference, connector-name dispatch, wildcard GitHub eligibility, generic SQL,
  caller-authored query, raw cursor, or automatic polling fallback.
- No replacement of PostgreSQL CDC or its explicit bootstrap handoff.
- No polling adapter for another database and no general-purpose transport-source generator.
- No unbounded in-memory collection: provider pages are emitted in bounded transport batches.

## Design

### Slice A — typed positive stream admission

Add `SourceStreamIneligibleError` at the registry boundary. `Registry.Preflight` still validates
the exact definition and mode before consulting the positive `eligible_streams` list; absence
returns the typed error before executor lookup or I/O. Wildcard semantics remain available only
for declarations that explicitly own `"*"`; GitHub uses the concrete stream list.

The GitHub list is generated/checked against the declarative bundle's executable streams. The
source executor becomes stream-neutral while retaining its exact definition-owned executor
reference. It validates that the selected stream is both declared eligible and present in the
engine connector before calling `engine.Connector.Read`.

`max_pages` is parsed only at this adapter boundary: omitted means one provider page; a positive
integer is a finite cap; `0`, `all`, or `unlimited` follows the stream's declared paginator to
exhaustion. Every provider callback is emitted as a bounded `SourcePage` with a stream/ordinal/
content-bound checkpoint candidate. An empty read emits no destination page and cannot advance a
persisted checkpoint.

### Slice B — PostgreSQL shared polling adapter

Promote `polling_watermark.json` only with an exact native source/apply executor pair and immutable
conformance evidence already understood by `PollingPreflight`. The authored declaration describes
the closed mechanism and bounds; a named PostgreSQL binder fills dynamic relation identity,
credential cohort, catalog fingerprint, caller-selected cursor column, and complete primary-key
tuple only after typed catalog discovery.

The native runner owns PostgreSQL concerns:

- read-only typed catalog and authorization admission;
- identifier-safe lexicographic SQL over `(cursor, primary-key...)`;
- lossless typed cursor and composite-key token encoding;
- page-size/request/time budgets and bounded pool use;
- non-null cursor/key checks and strict native traversal validation;
- schema/key/source-generation mismatch as typed rebootstrap outcomes.

The shared `engine.PollingSourceExecutor` remains the only page/checkpoint sequencer. The existing
`SnapshotTransportSource` outward executor reference stays definition-owned; it delegates only an
implemented, supported incremental polling request to the admitted shared executor. Full snapshot
and explicit bootstrap/CDC selection remain unchanged.

### Slice C — production spine conformance

Tests obtain connectors and the transport registry through `app.Open`/production composition.
They do not hand-register executors. The assertions follow:

`cmd/pm → cli.Run → app.Open → composeTransportRegistry → RegisterDeclaredTransports →`
`dispatchETLMode/runTransportETL → synctransport.Orchestrator → source executor → connection`
`warehouse Stage/Reopen → PostgreSQL managed target → durable receipt → checkpoint CAS`.

The production test matrix covers API→database and database→database routes. A declared-but-
ineligible probe proves the allowlist by returning the typed error with zero source requests,
staged pages, sends/rows, and checkpoint movement.

## TDD sequence

1. **Red A:** add registry and `app.Open` composition tests for GitHub `commits`, exact declaration
   coverage, multi-page bounded emission, and typed undeclared-stream zero-effect refusal.
2. **Green A:** implement the typed error, explicit GitHub stream list, neutral declarative source,
   `max_pages` semantics, and stream-bound candidates.
3. **Red B:** add PostgreSQL definition/native tests proving that production factory composition
   selects the shared polling executor, preserves `(cursor, complete PK tuple)`, and refuses unsafe
   shapes before any fetch/mutation.
4. **Green B:** implement the PostgreSQL declaration/binder/runner and delegate through
   `PollingSourceExecutor`; retain snapshot and CDC routes.
5. **Red/green C:** exercise cancellation, interruption/restart, empty/single/large, duplicate,
   out-of-order, schema drift, auth refusal, concurrency, resume, and acknowledged replay. Each
   refusal asserts its typed result and all zero/unchanged side effects.
6. **Live:** build `pm`; start the repository-owned PostgreSQL harness; run PostgreSQL polling and
   authenticated `rails/rails` commits with `max_pages=unlimited`; independently count warehouse
   and target rows. If infrastructure cannot start, record that fact and leave the live item open.
7. **Verify/review:** regenerate derived artifacts once, run focused tests/race tests and individual
   repository gates, then execute inline `verify-work` and `code-review`; close gaps before PR.

## Planned focused commands

```sh
go test -count=1 ./internal/synctransport ./internal/connectors/engine
go test -count=1 ./internal/connectors/native/postgres
go test -count=1 ./internal/app
go test -race -count=1 ./internal/synctransport ./internal/connectors/engine
go test -race -count=1 ./internal/connectors/native/postgres ./internal/app
go test -tags databaseintegration -count=1 ./internal/connectors/native/postgres
go test -count=1 ./internal/cli
go vet ./...
go build ./cmd/pm
```

The full `go test ./...` and monolithic `make verify` are deliberately left to CI. Local
verification runs the applicable non-suite make gates individually as required by `AGENTS.md`.

## Plan verification

- Every acceptance criterion has a red assertion before production edits and a focused green
  command in `TDD-LEDGER.md`.
- The design retains exact family/reference/evidence checks and adds a typed result only at the
  existing positive allowlist decision.
- Both requested production routes are composed from definitions; no test-only registration is an
  acceptance proof.
- Every edge named in the brief is assigned to a deterministic test or live check in
  `VERIFICATION.md`.
- Scope excludes the three prohibited issues and contains no dependency addition.
