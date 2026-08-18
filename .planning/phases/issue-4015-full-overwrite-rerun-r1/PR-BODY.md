Refs #4015

## Summary

- make saved-checkpoint eligibility follow source refresh semantics at the shared transport boundary
- force `full_append` and `full_overwrite` sources to start from origin on every run while preserving checkpoint resumption for all incremental modes
- use the same rule for generic, run-scoped overwrite, serial Arrow, and pipelined Arrow paths
- add live binary proofs for exact full-overwrite replacement and unchanged incremental-upsert skipping

## Diagnosis

The checkpoint-resumption hypothesis was confirmed. Before production edits, the six-mode boundary test failed only `full_append` and `full_overwrite`: both incorrectly received the prior checkpoint; all four incremental modes received it as intended.

The live PostgreSQL control reproduced the dangerous completed `0/0` second run. One symptom detail differed from the brief: this destination published an empty replacement instead of retaining stale rows. That is silent data loss, and it confirmed that destination publication was not the correct policy boundary.

## Red / Green / Refactor

| Stage | Evidence |
| --- | --- |
| Red | First run: `records_read=3`, `records_loaded=3`, target IDs `[1 2 3]`, sample `id=2 label="page-one-b"`. Changed source IDs `[2 3 4]`, sample `replacement-two`. Second run reported completed `0/0`; independent target query returned zero rows. |
| Green | The same rerun transferred `3/3`; independent target query returned exactly IDs `[2 3 4]`, sample `id=2 label="replacement-two"`, and zero rows for removed ID 1. |
| Incremental | Dedicated unchanged `incremental_upsert`: first run `3/3`; rerun `0/0`; target stayed at three rows and sample `id=2 label="event-two"`. |
| Refactor | Replaced two Arrow-only guards plus both generic source request sites with one `sourceCheckpointForMode` rule. |

## Mode matrix

| Mode | Source behavior after this change | Destination behavior |
| --- | --- | --- |
| `full_append` | full reread; prior checkpoint omitted | append unchanged |
| `full_overwrite` | full reread; prior checkpoint omitted | one run-scoped replacement unchanged |
| `incremental_append` | resume from prior checkpoint | append unchanged |
| `incremental_dedupe` | resume from prior checkpoint | dedupe unchanged |
| `incremental_dedupe_history` | resume from prior checkpoint | history dedupe unchanged |
| `incremental_upsert` | resume from prior checkpoint | merge/upsert unchanged |

`Resume` identity/generation, destination planning, durable read-back, final candidate selection, downstream acknowledgement, and checkpoint commit ordering are unchanged.

## GSD lifecycle

- resolved and executed the repository-local `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts
- recorded the required inline/manual fallback because this dispatched slice is not a numbered roadmap phase and the canonical contract prohibits spawning reviewer/planner roles
- committed planning before production edits, then separate red, green, and verification checkpoints
- durable evidence: `.planning/phases/issue-4015-full-overwrite-rerun-r1/`

## Required skills used

- `golang-how-to`
- `golang-troubleshooting`
- `golang-testing`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-context`
- `golang-concurrency`
- `golang-database`
- `golang-lint`

The runtime/PostgreSQL integration reference and database harness contract were also applied. No CLI surface changed, so CLI help/manual/website parity is not applicable.

## Verification

Passed locally:

```text
go test -count=1 -timeout 20m ./internal/synctransport
go test -count=1 -timeout 20m ./internal/app
go test -count=1 -timeout 20m ./internal/cli
go vet ./...
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
scripts/verify-gsd-workflow
```

The two opt-in live commands and exact output are recorded in `traces/green-full-refresh-checkpoint.md`. Full-tree `go test ./...` and monolithic `make verify` remain CI-owned per `AGENTS.md`; local runs used the required scoped-package and individual-gate strategy.

## Safety and scope

- no secret value was requested, printed, logged, or persisted
- used only the explicit existing local Docker Unix endpoint; did not restart Colima or Docker
- harness resources remained uniquely owned with unconditional cleanup
- no dependency, connector surface, CLI output, release target, or documentation contract changed
- did not touch `release/0.2.0-mvp` or PR #4250
- unrelated PostgreSQL CDC restart recovery remains outside scope in PR #4253

## Commit checkpoints

- `3b7707a4d` — GSD context and TDD plan
- `d24833287` — red unit/live reproduction
- `2b1bff0e7` — shared boundary fix and green proofs
- `aefd515a0` — verification, UAT, and inline review evidence

## Review route

- local GSD review: clean; no unresolved critical, warning, or informational finding
- primary automated route: `claude_auto` on non-draft PR open
- primary status: pending PR creation
- coverage route: `sub_pr`
- fallback: none; Copilot was not requested
- human gate: parent/default-branch merge remains human-controlled
