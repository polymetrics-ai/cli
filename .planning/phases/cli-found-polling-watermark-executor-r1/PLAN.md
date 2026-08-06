# PLAN — polling-watermark changefeed executor

## Scope

Implement issue #3855 on branch `fm/cli-found-polling-watermark-executor-r1` as
one shared, declaration-driven polling-watermark `ChangefeedExecutor`.

Owned paths:

- `internal/connectors/**` and `internal/connectors/engine/**` for the closed
  declaration, executor, fail-closed integration, and unit tests.
- `docs/migration/conventions.md` for the connector-authoring contract.
- `.planning/phases/cli-found-polling-watermark-executor-r1/**` for the
  required GSD/TDD evidence.

The only connector declaration added by this branch is an engine test fixture
under `internal/connectors/engine/testdata/`. No production connector is
claimed CDC-capable. PostgreSQL remains `logical_replication: unsupported`.

Out of scope:

- webhook and event-stream implementations;
- provider calls, credentials, live databases, dependencies, or per-connector
  Go code;
- a generic SQL, HTTP-write, shell, or raw-protocol surface;
- weakening `HasImplementedChangefeed`, catalog/list/manifest/definition CDC
  projection, or the implemented-command runtime-preflight sweep;
- replacing the #3810 durable database-sync contract.

## GSD lifecycle and skills

`scripts/gsd doctor`, all five required `scripts/gsd sources` resolutions, and
`go run ./cmd/agentcontractgen check` passed. Generated prompts for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` were inspected. This isolated worker has no compatible Pi role
runtime and the canonical contract forbids role spawning, so the generated
workflow is executed inline with these durable artifacts. This is the approved
manual fallback and does not weaken TDD, verification, or review gates.

Loaded per the required-skills routing:
`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`,
`golang-context`, `golang-concurrency`, `golang-database`, `golang-cli`, and
`golang-documentation`.

The CLI help/docs/website parity reference was read because an implemented
changefeed changes connector capability projection. No command, flag, help,
manual, or website page is expected to change; the verification record will
explicitly confirm this is still true.

## Contract and compatibility decisions

1. A polling declaration supplies its provider ordering value kind and record
   extraction path, stable tie-breaker extraction path, inclusive boundary
   policy, optional timestamp safety lag, page size, maximum pages, and request
   budget. The only supported boundary policy is re-read from `>= checkpoint`.
   It intentionally produces duplicates and therefore requires
   `delivery.duplicates: at_least_once`.
2. The executor advances after a destination `Accept` returns successfully,
   then persists the resulting checkpoint through the small
   `ChangefeedCheckpointCommitter`
   interface. A process crash before persistence must replay the last page;
   the required destination contract is consequently idempotent/replay-safe.
3. Timestamp watermark declarations explicitly set `safety_lag_seconds`.
   Positive values make the executor subtract the lag from a committed
   timestamp when constructing the next-run overlap boundary; an initial
   checkpoint deliberately leaves the source's declared snapshot/barrier
   choice untouched. `0` is an explicit opt-out that is documented as capable
   of losing late arrivals. No wall-clock is read directly in tests. Opaque
   cursor and monotonic-sequence declarations must set this field to `0`
   rather than pretend to use timestamp lag.
4. Hard deletes are not observable. A declaration may only advertise deletes
   when it names a soft-delete extraction path or a deletion-feed declaration;
   otherwise it must say `not_available`. Soft deletes are emitted as tombstone
   events only when the source record supplies the declared marker.
5. #3810 is open. This branch consumes no app state or scalar cursor. Its
   narrow checkpoint store accepts an opaque committed value and can be adapted
   to #3810's versioned checkpoint envelope as a substitution, not a rewrite.

## TDD sequence

1. **Red: declaration and fail-closed promotion.** Add a test-only bundle
   declaring polling-watermark and prove it is not CDC-capable until its
   registered executor fully matches the checkpoint contract. Prove each
   missing required checkpoint field remains rejected.
2. **Red: execution semantics.** Write executor tests for equal-watermark
   records over a page edge, lagged timestamp bounds from a committed checkpoint,
   soft-delete versus hard-delete visibility, cancelled context, and every
   declared work bound.
3. **Red: durability ordering.** Add a destination/store test double that
   fails after accepted records but before checkpoint persistence. A rerun must
   replay the window and never advance past it; a destination failure must not
   store a checkpoint.
4. **Green: smallest shared executor.** Parse/validate the declaration in the
   existing engine loader, create the executor from only that declaration, and
   keep source fetch/destination acceptance/store persistence injectable.
5. **Green: real gate.** Register the executor path so the existing
   `HasImplementedChangefeed` capability projection derives CDC only through
   the full matching descriptor. Do not add a parallel validator which could
   drift from runtime preflight.
6. **Refactor and document.** Keep the public structs closed and bounded,
   explain ties/lag/delete truth in the migration convention, then execute the
   focused matrix and individual repository gates.

## Expected implementation seams (to confirm after red tests)

- `internal/connectors/connectors.go` owns public descriptor semantics and the
  matching executor gate.
- `internal/connectors/engine/bundle.go` owns loading a test bundle's
  `changefeed.json`; its schema is `engine/schema/changefeed.schema.json`.
- `internal/connectors/engine` is the shared declarative execution owner; its
  transport callback is a narrow read-only page interface, not a provider API.
- The destination acknowledgement and persistence port are explicit local
  interfaces, making #3810 envelope adoption a constructor/store adapter
  change.

## Commit checkpoints

1. Plan/TDD/verification artifact checkpoint (this commit).
2. Red tests plus test-only bundle.
3. Shared executor and green focused tests.
4. Documentation and verification/review fixes.
