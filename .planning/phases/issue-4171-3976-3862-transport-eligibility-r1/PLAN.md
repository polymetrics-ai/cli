# PLAN — transport source eligibility club (#4171, #3862; #3976 deconflicted)

## Scope correction — 2026-08-16

PR 4175 owns #3976's PostgreSQL resumable-read implementation. A binary-level review of this
branch found no standalone production preflight that can bind PostgreSQL's dynamic catalog object:
the attempted bind occurs only inside source read execution after authentication admission and
typed-catalog I/O. Therefore this branch must not advertise `polling_watermark` as implemented or
duplicate PR 4175's adapter. The declaration remains `planned` with its blocking reason, while this
PR delivers #4171 and #3862 only. This is a deliberate deconfliction, not a relaxed test.

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
2. PostgreSQL's `polling_watermark` declaration remains honestly `planned` until a shipped
   production preflight can bind its dynamic source/object/destination contract; #3976's adapter
   implementation is owned by PR 4175.
3. Cross-family conformance is proven from the shipped entry path, while undeclared streams,
   executors, and evidence remain typed pre-I/O refusals.

## Scope and ownership

- `internal/connectors/defs/github/**`: explicit source eligibility and derived bundle artifacts.
- `internal/app/**`: production composition for the general declarative source and focused
  shipped-entry tests.
- `internal/synctransport/**`: typed stream-ineligibility outcome without changing admission
  order, exact executor lookup, or definition-owned evidence verification.
- `internal/connectors/defs/postgres/polling_watermark.json`: retain the planned declaration and
  reason; do not add a source/apply runtime contract or adapter in this branch.
- This phase directory: red/green traces, verification, review, summary, and PR evidence.

No CLI command, flag, or output shape is added. `max_pages` remains a connector runtime-config
key, so CLI help/manual/website parity is not applicable; connector authoring and operator docs are
updated only if the existing generated bundle/docs surface changes.

## Non-goals

- No edits for #4125, #4158, or #4169.
- No capability-bit inference, connector-name dispatch, wildcard GitHub eligibility, generic SQL,
  caller-authored query, raw cursor, or automatic polling fallback.
- No replacement of PostgreSQL CDC or its explicit bootstrap handoff.
- No #3976 polling implementation: it is deconflicted to PR 4175. This branch must neither
  duplicate nor pre-empt that work.
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

### Slice B — PostgreSQL polling truthfulness and deconfliction

Keep the existing declaration `planned`. `app.Open` may compose the closed snapshot transport, but
that is not polling preflight: dynamic source/object/destination binding currently appears only in
the attempted `ReadTransport` path after authentication and catalog I/O. This branch has no
shipped-binary proof of a successful polling preflight and therefore leaves both the declaration
and production path unchanged. PR 4175 owns the exact native/shared-poller implementation.

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
3. **Red B:** `TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt` must fail
   when inspection advertises `implemented` without a bindable production preflight.
4. **Green B:** restore the planned declaration and its blocking reason; do not change the guard or
   add a test-only binding.
5. **Red/green C:** exercise cancellation, interruption/restart, empty/single/large, duplicate,
   out-of-order, schema drift, auth refusal, concurrency, resume, and acknowledged replay. Each
   refusal asserts its typed result and all zero/unchanged side effects.
6. **Live:** build `pm`; when the credential/runtime limits clear, run authenticated `rails/rails`
   commits with `max_pages=unlimited` and independently count warehouse and target rows. PostgreSQL
   polling execution is deliberately deferred to PR 4175. If infrastructure cannot start, record
   that fact and leave the GitHub live item open.
7. **Verify/review:** regenerate derived artifacts once, run focused tests/race tests and individual
   repository gates, then execute inline `verify-work` and `code-review`; close gaps before PR.

## Planned focused commands

```sh
go test -count=1 ./internal/synctransport ./internal/connectors/engine
go test -count=1 ./internal/app
go test -race -count=1 ./internal/synctransport ./internal/connectors/engine
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
- GitHub's production route is composed from definitions; no test-only registration is an
  acceptance proof. PostgreSQL polling remains planned until PR 4175 can deliver its own proof.
- Every edge named in the brief is assigned to a deterministic test or live check in
  `VERIFICATION.md`.
- Scope excludes the three prohibited issues and contains no dependency addition.
