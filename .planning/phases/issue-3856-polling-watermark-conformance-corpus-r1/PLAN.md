# PLAN — #3856 immutable polling-watermark conformance corpus

## Scope and ownership

Create one embedded `v1` polling corpus and a reusable runner under the common
engine polling area, plus focused tests and this phase evidence. The owned paths
are expected to be:

- `internal/connectors/engine/polling_conformance.go`
- `internal/connectors/engine/testdata/polling_watermark_conformance/v1.json`
- `internal/connectors/engine/polling_conformance_test.go`
- `.planning/phases/issue-3856-polling-watermark-conformance-corpus-r1/**`

No connector definition, `cmd/`, `internal/app/` production path, public CDC
projection, `synccontract` generic fixture, provider, credential, database, or
apply strategy is in scope.

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-testing`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`, and `golang-lint`.

The lifecycle is inline/manual execution of generated GSD prompts because the
canonical single-worker contract forbids GSD-role spawning:
`discuss-phase --auto` -> `plan-phase --tdd --auto` -> `execute-phase --auto`
-> `verify-work --auto` -> `code-review --auto`.

## TDD plan

1. **RED — runner contract.** Add
   `TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture` before the
   loader or runner exists. Its deterministic lane factory must split equal
   watermarks across physical pages, acknowledge a page then fail persistence,
   restart from the prior committed envelope, retain every stable identity, and
   assert the complete immutable ID set ran. The proposed symbols must not yet
   compile.
2. **GREEN — immutable corpus and loader.** Embed a standalone versioned JSON
   corpus, validate unique IDs and all required scenario kinds, expose only
   defensive copies and digest/evidence derived from its exact bytes, and never
   mutate #3810's corpus.
3. **GREEN — registered no-skip runner.** Accept only a registered polling
   executor/descriptor with exact evidence. The runner owns fixture enumeration
   and executes every scenario. No public filter, skip, replacement, or
   lane-provided list is exposed.
4. **GREEN — reference behavior.** Use a deterministic fake lane to prove each
   corpus scenario asserts concrete source page requests, durable
   acknowledgements, persisted envelope transitions, recovery outcomes,
   tombstone/history state, and admission failure rather than merely matching
   fixture strings.
5. **REFACTOR.** Clarify cloning/validation boundaries and package comments;
   retain only narrow test-only seam interfaces, no goroutine or external I/O.

## Required corpus ledger

| ID | Required behavior |
| --- | --- |
| `equal-watermark-page-split-recovery` | Composite tuple replay across physical pages and crash recovery has no stable-identity loss. |
| `empty-page-does-not-advance` | Empty read leaves the committed envelope unchanged. |
| `non-advancing-page-is-rejected` | Repeated cursor/page continuation is refused. |
| `null-precision-coercion-is-rejected` | NULL or lossy/ambiguous cursor conversion fails before scan. |
| `unstable-or-non-unique-keyset-is-rejected` | Missing stable unique tie-breaker is refused. |
| `bounded-overlap-and-commit-lag` | Bounded overlap plus ordering fence permits replay; unbounded lag is refused. |
| `source-generation-or-schema-mismatch` | Resume produces a typed rebootstrap outcome and never clears state. |
| `acknowledged-before-checkpoint-replays` | Ack followed by failed persistence replays from prior committed envelope. |
| `tombstone-history-and-hard-delete-visibility` | Soft delete closes history/tombstones; hard delete remains invisible. |
| `missing-executor-or-evidence-is-rejected` | Unregistered executor or incorrect evidence cannot enter the suite. |

## Verification plan

- `go test -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformance' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/synccontract -count=1`
- `go test -timeout 20m ./internal/app -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`
- relevant `-race` engine package run; `gofmt`, `go vet`, build, and individual
  repository gates from `AGENTS.md`.

No CLI surface changes are planned, so help/manual/website parity is explicitly
not applicable after the diff is checked. No live or credential-backed test is
authorized.

## Checkpoints and correction ledger

1. Plan/TDD phase artifact checkpoint.
2. First RED test checkpoint (no production implementation).
3. GREEN corpus/runner/reference-lane checkpoint.
4. Verification/review-fix checkpoint if needed.

Correction rounds begin at **0/5**. A round increments only if a gate finds a
production/test defect and the candidate changes in response.
