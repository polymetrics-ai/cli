---
phase: issue-4015-full-overwrite-rerun-r1
plan: 01
type: tdd
wave: 1
depends_on: []
files_modified:
  - internal/synctransport/orchestrator.go
  - internal/synctransport/arrow_fast_path_controller.go
  - internal/synctransport/arrow_fast_path_pipeline.go
  - internal/synctransport/transport_test.go
  - internal/cli/postgres_transport_binary_integration_test.go
  - .planning/phases/issue-4015-full-overwrite-rerun-r1/*
autonomous: true
requirements:
  - REFS-4015-FULL-OVERWRITE-RERUN
---

# Issue 4015 full-overwrite re-run correctness — Plan

## Objective

Make checkpoint eligibility follow the source-refresh contract: full-refresh modes start at the beginning on every run, while incremental modes resume from their committed position. Prove a second real-binary full-overwrite run replaces the target after source changes and that unchanged incremental input still skips.

## Lifecycle

- `scripts/gsd prompt discuss-phase issue-4015-full-overwrite-rerun-r1 --auto`
- `scripts/gsd prompt plan-phase issue-4015-full-overwrite-rerun-r1 --tdd --skip-research --auto`
- `scripts/gsd prompt execute-phase issue-4015-full-overwrite-rerun-r1 --interactive --auto`
- `scripts/gsd prompt verify-work issue-4015-full-overwrite-rerun-r1 --auto`
- `scripts/gsd prompt code-review issue-4015-full-overwrite-rerun-r1 --depth=deep --files=internal/synctransport/orchestrator.go,internal/synctransport/arrow_fast_path_controller.go,internal/synctransport/arrow_fast_path_pipeline.go,internal/synctransport/transport_test.go,internal/cli/postgres_transport_binary_integration_test.go`

The generated prompts are executed inline because this dispatched bug is not a numbered roadmap phase and the repository's canonical single-worker contract forbids spawning delivery roles.

## Required skills

- `golang-how-to`
- `golang-troubleshooting`
- `golang-testing`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-database`
- GSD lifecycle skills: `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`

## Threat model

| Threat | Severity | Boundary | Mitigation and evidence |
| --- | --- | --- | --- |
| A stale checkpoint suppresses a full refresh and leaves stale destination data while the run reports success | high | Persisted stream state → source extraction | Centralize mode-based checkpoint eligibility and test all six requested modes at the source request boundary. |
| A broad fix disables incremental resume and duplicates or reprocesses data | high | Mode contract → checkpoint selection | Table-driven regression asserts every incremental mode receives the prior checkpoint; live `incremental_upsert` replay remains `0/0`. |
| A destructive overwrite publishes partial pages | high | Source pages → destination publication | Preserve the existing run-scoped `Begin/Apply/Publish/ReadBack` protocol and its one final acknowledged checkpoint. |
| Evidence leaks credentials or relies on a shared runtime default | high | Test harness → local runtime | Use only generated PostgreSQL fixture credentials and the explicit direct Unix container endpoint; do not print secret values or restart runtimes. |

No new dependency, schema migration, generic SQL surface, auth change, or public action is introduced.

<feature>
  <name>Source-semantic checkpoint eligibility</name>
  <files>internal/synctransport/orchestrator.go, internal/synctransport/arrow_fast_path_controller.go, internal/synctransport/arrow_fast_path_pipeline.go, internal/synctransport/transport_test.go, internal/cli/postgres_transport_binary_integration_test.go</files>
  <behavior>
    `full_append` and `full_overwrite` source requests receive no prior checkpoint. `incremental_append`, `incremental_dedupe`, `incremental_dedupe_history`, and `incremental_upsert` receive the committed checkpoint unchanged. A second full-overwrite run after deleting, updating, and adding source rows reports the complete current source count and replaces destination rows exactly.
  </behavior>
  <implementation>Introduce one transport-level checkpoint-selection rule derived from the canonical sync mode and use it for regular, run-scoped overwrite, serial Arrow, and pipelined Arrow source requests. Do not alter resume identity, checkpoint commit ordering, destination publication, or incremental source behavior.</implementation>
</feature>

<tasks>
  <task type="tdd" gate="red">
    <name>Reproduce the stale full-overwrite re-run and pin the mode matrix</name>
    <read_first>CONTEXT.md; internal/app/sync_modes.go; internal/synctransport/orchestrator.go; internal/synctransport/types.go; internal/synctransport/transport_test.go; internal/connectors/engine/polling_source.go; internal/cli/postgres_transport_binary_integration_test.go; internal/connectors/native/dbtest/README.md</read_first>
    <action>Add a table-driven orchestrator test covering the six requested canonical modes and asserting whether the source receives the prior checkpoint. Extend the real PostgreSQL binary full-overwrite test to perform run one, independently read the target, delete source ID 1, update source ID 2 to a named v2 sample, insert source ID 4, run again, then independently assert target IDs `[2,3,4]`, ID 1 absence, and the v2 sample. Run both before production edits and capture exact failures/counts in `traces/`.</action>
    <verify>`go test -count=1 -run '^TestOrchestratorSourceCheckpointFollowsRefreshSemantics$' ./internal/synctransport` fails because full modes still receive the prior checkpoint; opt-in `TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage` fails on run two with exact observed counts and stale target state.</verify>
    <acceptance_criteria>The unit red is an assertion failure, not a compile failure. The live red records first-run counts, second-run `records_read`/`records_loaded`, target count, named sample, and presence/absence of IDs 1 and 4.</acceptance_criteria>
  </task>

  <task type="tdd" gate="green">
    <name>Make source checkpoint selection follow the canonical mode</name>
    <read_first>the failing tests; internal/synctransport/orchestrator.go; internal/synctransport/arrow_fast_path_controller.go; internal/synctransport/arrow_fast_path_pipeline.go; internal/synccontract/mode.go</read_first>
    <action>Add one unexported mode-based checkpoint selector in `synctransport`: return `nil` for `ModeFullAppend` and `ModeFullOverwrite`; return the prior checkpoint for all incremental modes and `ModeChangeCapture`. Route all source request construction through it, replacing the two Arrow-local full-overwrite conditions. Keep `Resume`, destination plan/publication, and commit callbacks unchanged.</action>
    <verify>The unit mode matrix passes, the live full-overwrite test reports the exact complete second-run count and independently proves replacement, and the live incremental-upsert replay still reports `0/0` with unchanged target content.</verify>
    <acceptance_criteria>No provider-specific conditional is added. Both full-refresh modes ignore prior checkpoints. Every incremental mode preserves its checkpoint. All four source-extraction paths use the same rule.</acceptance_criteria>
  </task>

  <task type="tdd" gate="refactor">
    <name>Consolidate evidence and run repository gates</name>
    <read_first>all changed files; AGENTS.md verification section; VERIFICATION.md; TDD-LEDGER.md</read_first>
    <action>Run gofmt and targeted tests, then vet/build and every applicable non-monolithic `make verify` gate. Record exact commands/results and any skipped full-suite command with the repository-prescribed reason. Update `SUMMARY.md`, `VERIFICATION.md`, `REVIEW.md`, and PR body evidence.</action>
    <verify>`git diff --check`, targeted tests, vet/build, generated-file checks, and review are green.</verify>
    <acceptance_criteria>Red and green evidence includes exact observable counts; no secret value or runtime-default dependency is recorded; the diff is limited to the shared checkpoint boundary, regression tests, and planning evidence.</acceptance_criteria>
  </task>
</tasks>

## Verification

- Unit mode matrix and focused `internal/synctransport` package tests.
- Opt-in real-binary PostgreSQL full-overwrite rerun and existing incremental-upsert replay proof through the explicit local container endpoint.
- Changed-package tests with `-timeout 20m`, plus `internal/cli` separately.
- `gofmt`, `go vet ./...`, `go build ./cmd/pm`, and applicable generated/snapshot/contract gates.
- Deep code review of every changed production and test file.

## Success criteria

- The red live run reproduces exact `0/0` and independently observes stale target state before the fix.
- The green live run reports the complete changed-source count and independently observes exact replacement, including deleted-row absence.
- `full_append` also re-reads from the start by contract.
- Every incremental mode retains checkpoint resume, and live `incremental_upsert` unchanged replay remains `0/0`.
- Required commits, push, direct PR, API base verification, and PR review routing are complete.

