# Issue 372: GSD Pi + Go Shepherd

## Objective

Adopt pinned `@opengsd/gsd-pi@1.11.0` as the workflow engine and add a standalone Go governance
module without coupling it to the Polymetrics CLI module.

## Delivery slices

1. Qualify supported GSD Pi headless interfaces and record incompatibilities (#373).
2. Encode the Polymetrics issue-first/TDD/review/human-gate overlay (#374).
3. Build the typed Go governance domain and SQLite authority store (#375).
4. Add bounded headless execution, allowlisted event projection, liveness, and query reconciliation
   (#376).
5. Add capability grants, fencing, exact-head ratification, and an idempotent GitHub outbox (#377).
6. Add redacted local telemetry and analytics export boundaries (#378).
7. Replay known incidents and run a merge-disabled canary before legacy removal (#379).

## Boundaries

- GSD Pi owns workflow state. Shepherd only uses documented CLI/query/event surfaces.
- Shepherd owns external-effect authorization and never autonomously merges to `main`.
- No raw prompts, reasoning, credentials, command arguments, or tool results are persisted.
- The module lives under `agent-runtime/shepherd/` with a separate `go.mod`.
- Legacy removal and issue/PR closure are a final, separately reviewed cutover step.

## Required skills and workflows

- `golang-how-to`, `go-engineering`, Go project-layout, design-patterns, observability,
  concurrency, context, testing, error-handling, safety, and database guidance.
- `gsd-programming-loop` with RED/GREEN/refactor evidence.
- Issue-first and parent/subissue orchestration contracts.

## Execution decisions

- Qualification: `local_critical_path` because it blocks every implementation slice.
- Architecture/workflow review: `read_only_spawned` using two non-mutating review agents.
- Production implementation: `local_critical_path` in the dedicated issue branch; no shared-checkout
  mutating subagent is permitted.

