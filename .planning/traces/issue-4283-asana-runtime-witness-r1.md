# Issue #4283: Asana runtime witness harness R1

## Task Delivery Header

- Issue: Refs #4283 — Batch-1 source-rigidity R2, Asana witness-harness slice.
- Base branch: `origin/fm/cli-top100-declaration-batch-r1` at `f38c205b052c7b1be281b83d70e0e78668c26835`.
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main`.
- Delivery: one normal committed and pushed candidate branch for parent review/integration; no merge.
- Working branch: `codex/4283-asana-runtime-witness-r1`.
- Task: Add test-only local `httptest` plus temporary-DuckDB witnesses for the already-declared twelve Asana ETL streams and the declared project/task event transport. The tests must not expand source coverage, change runtime behavior, or promote the 52 mapped-unproven ETL cells.
- Verification: focused red/green tests, `gofmt`, affected-package tests, focused race test where feasible, `go vet`, `agentcontractgen check`, `git diff --check`, and a changed-path audit.
- Skills: `connector-lane-build-order`, `go-engineering`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-database`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every existing declared Asana ETL stream reaches the saved provider-to-DuckDB path | fake | A local `httptest` provider is necessary because the task has no authorized Asana sandbox or credentials. The test will assert each named stream performs a paginated read and durably materializes its expected row count in its connection-owned DuckDB warehouse. |
| A declared Asana ETL non-success does not materialize a partial warehouse result | fake | The local provider double returns a declared non-success; the test asserts `RunETL` returns an error and the target table has zero rows. |
| Declared Asana project/task event transport bootstraps, replays a durable token window, and records a tombstone through DuckDB | fake | A local event executor/provider double is necessary because no provider-live event subscription is authorized. The test will assert stored rows and durable checkpoint/token behavior, including delete/tombstone semantics, rather than merely checking for no error. |

## Manual GSD/TDD Fallback

`scripts/gsd doctor`, `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`, generated prompts for all five lifecycle stages, and `go run ./cmd/agentcontractgen check` passed before implementation. The assigned isolated worker cannot invoke the repository Pi role runtime without spawning lifecycle roles; inline execution is the documented fallback. This trace is the durable discuss/plan/execute/verify/review record.

## Plan

1. Reuse the existing CLI/App test harness patterns; add no production seams, dependencies, or source-definition artifacts.
2. Add one focused ETL witness test covering exactly the twelve currently implemented stream declarations, a multi-page response, and one non-success boundary.
3. Add one focused App-level event-transport witness test covering declared project/task event source bootstrap, replay, tombstone, token/checkpoint durability, and DuckDB materialization.
4. Record the red result before the test-only implementation, then run focused green and boundary checks.

## Red

- `go test -count=1 -timeout=20m ./internal/cli -run '^TestAsanaDeclaredETL(Stream|Streams)'` — red as intended: the new executable assertions failed to compile because `newAsanaETLFixture` did not yet exist.
- `go test -count=1 -timeout=20m ./internal/app -run '^TestAsanaEventTransportMaterializesBootstrapResumeAndTombstone$'` — red as intended: the app-level production-route assertion failed to compile because `newAsanaEventTransportHTTPFixture` did not yet exist.
- First runtime pass after adding the ETL fixture correctly refused the closed `full_append` form before provider I/O because no Asana full-snapshot transport has externally verified conformance. The test was narrowed to the pre-existing `full_refresh_append` compatibility route; it does not assert closed transport admission.
- The next runtime pass correctly refused the test server `base_url` because every selected Asana source-bound stream owns its provider origin. The fixture now preserves the declared `https://app.asana.com/api/1.0` URL and redirects only its TLS dial to local `httptest`; this is a test harness correction, not a runtime relaxation.
- No connector/runtime source change is expected or authorized: existing declarations and production executors are the subject under test.

## Green

- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go test -count=1 -timeout=20m ./internal/cli -run '^TestAsanaDeclaredETL(Stream|Streams)'` — passed: every existing declared twelve-stream collection route completed through `App.RunETL` to its temporary DuckDB table; each fixture response required an actual `next_page` continuation, while the non-success route stayed terminal with no warehouse rows. The fixture accepts each source-declared query shape rather than incorrectly requiring `limit=100` for every provider route, and it confirms any retry remains the same first-page task route.
- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go test -count=1 -timeout=20m ./internal/app -run '^TestAsanaEventTransportMaterializesBootstrapResumeAndTombstone$'` — passed: initial 412 token bootstrap, reopen/resume, multi-page token exhaustion, one hydrated task, one durable deleted-task tombstone, relationship-removal non-tombstone behavior, final empty replay, and persisted checkpoint tokens all traverse the existing app transport into temporary DuckDB. The resumed run records one loaded row because the delete is carried as a tombstone and is separately asserted from DuckDB.
- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go test -race -count=1 -timeout=20m ./internal/cli -run '^TestAsanaDeclaredETL(Stream|Streams)'` — passed in 35.293s.
- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go test -race -count=1 -timeout=20m ./internal/app -run '^TestAsanaEventTransportMaterializesBootstrapResumeAndTombstone$'` — passed in 39.714s.
- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go vet ./internal/cli ./internal/app` — passed.
- `GOCACHE=/private/tmp/cli-4283-asana-runtime-witness-r1/.gocache-asana-r1 go run ./cmd/agentcontractgen check` — passed: canonical contract and registered projections are current.
- `gofmt -d` on both test files and `git diff --check` — clean. The task-isolated build cache remains untracked and was not cleaned.

## Non-goals

- No production runtime, command surface, source lock, source matrix, enabled contract, generated artifact, Foundation Atlas, or provider-live/credential change.
- The 52 Asana ETL `mapped_unproven` cells remain mapped-unproven; this slice witnesses only the existing twelve declared streams.
