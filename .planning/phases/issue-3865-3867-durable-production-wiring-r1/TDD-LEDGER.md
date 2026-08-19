# TDD LEDGER — #3865/#3867 durable production wiring

## Red

- `go test ./internal/coordination -run 'TestDurableCoordinationStores|TestFileCoordinationStores' -count=1`
  fails to compile at `durable_store_test.go` because
  `OpenFileAuthCohortHealthStore`, `OpenFileRateParkingStore`, and
  `ErrCoordinationStoreSchema` do not exist. The test starts a writer process,
  waits until both transitions are durable, kills that process, and requires a
  newly started process to observe the fence and resume the parked checkpoint.
- The first live production auth run reached `pm etl check` but failed with
  `postgres connector requires config host`: the command surface discarded its
  selected credential/runtime and therefore could not reach the durable
  admission path. The production test also specified that a runtime resolved
  before repair must retain its old epoch rather than silently joining repair.
- The first production parking composition used a test connector that consulted
  parking admission itself. Review rejected that evidence shape because it did
  not prove the registered declarative engine send path; the final test uses an
  engine bundle and real HTTP server through `cli.Run`/`app.Open` instead.
- PR #4164 CI reproduced a slow-runner boundary in
  `TestProductionParkingCompositionSurvivesProcessKill`: the fixture's
  two-second provider reset elapsed while the restarted process was opening, so
  recovery correctly resumed before the test could assert the pre-reset fence.
  A local three-second post-crash delay reproduced the same
  `same-scope admission error = <nil>` failure deterministically.

## Green

- Versioned JSON auth/parking stores use the product's existing atomic
  replace/fsync state layer plus a kernel-released advisory file lock. A killed
  writer process leaves both files reopenable; a later process observes the
  fence and resumes the exact parked committed checkpoint.
- Auth repair publishes a new durable epoch. The three captain-directed cases
  are green: unrepaired callers receive `ErrAuthCohortFenced`; a fresh
  post-repair runtime reaches real PostgreSQL and the durable session counter
  advances; a pre-repair runtime receives `ErrAuthCohortEpochMismatch` with
  zero PostgreSQL sessions and no checkpoint mutation.
- Parking is constructed by `app.Open`, attached during credential resolution,
  admitted at engine request send, persisted only from typed terminal 429 reset
  evidence, and reloaded/claimed/resumed through `App.RunETL`. Three real child
  processes prove zero HTTP sends/checkpoint movement while parked and exactly
  one send after restart resume.
- Edge tests cover cancellation, SIGKILL/advisory-lock recovery, empty/single/
  128-record stores, idempotent duplicates, stale CAS/claim completion,
  checkpoint/schema drift, filesystem and protocol authentication refusal,
  concurrent same-scope admission/claim renewal, active-claim cancellation
  refusal, interrupted resume, and already-acknowledged replay. Two restarted
  CLI processes race the due record and produce exactly one HTTP resume send.
- The process-kill parking fixture now persists a thirty-second provider reset,
  reopens the durable store after killing the writer, asserts the exact run and
  reset boundary, then launches production admission before that boundary and
  production recovery after it. The deterministic three-second delay remains
  as regression coverage for the CI timing that defeated the old fixture.
- Green commands:
  - `go test -race -timeout 20m ./internal/coordination`
  - `go test -timeout 20m ./internal/app`
  - `go test -timeout 20m ./internal/connectors/engine`
  - `go test -timeout 20m ./internal/connectors/native/postgres`
  - `go test -timeout 20m ./internal/cli -run '^TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess$' -count=1`
  - `go test -timeout 20m ./internal/app -run '^TestProductionParkingCompositionSurvivesProcessKill$' -count=1`
  - `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -timeout 20m ./internal/cli -run '^TestCLIDurableAuthenticationFenceAndRepairLivePostgres$' -count=1 -v`

## Refactor

- Replaced a manual parking fixture with an engine-backed connector and local
  HTTP server so the production registry/dispatch/admission call chain is the
  thing under test.
- Moved authentication reporting to connector-owned operation boundaries so a
  verified source failure cannot fence an unrelated destination. Engine error
  maps classify only matching declared rules; PostgreSQL classifies only
  SQLSTATE `28P01`, not permission or transport errors.
- Rebased onto the #4165 integration-base update, regenerated surface/docs/
  skills/golden/website projections in one pass, and confirmed that the pass
  changed no derived artifact. Focused app/CLI/certify/coordination tests plus
  vet, lint, boundary, catalog, release, and GSD drift gates are green.

## GSD lifecycle

- `scripts/gsd prompt discuss-phase 3865-3867-durable-coordination-r1 --auto`
- `scripts/gsd prompt plan-phase 3865-3867-durable-coordination-r1 --tdd --skip-research`
- `scripts/gsd sources execute-phase`
- `scripts/gsd prompt execute-phase 3865-3867-durable-coordination-r1`
- `scripts/gsd prompt plan-phase 3865-3867-durable-coordination-r1 --gaps --tdd --skip-research`
- `scripts/gsd prompt execute-phase 3865-3867-durable-coordination-r1 --gaps-only`
- `scripts/gsd prompt verify-work 3865-3867-durable-coordination-r1`
- `scripts/gsd prompt code-review 3865-3867-durable-coordination-r1`
- Inline/manual fallback: issue-scoped work is not a numbered roadmap phase and
  the canonical contract forbids role spawning.
