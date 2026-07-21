# PRD Coverage

Phase: `471-pi-agent-session-shepherd`

| Source | Coverage |
|---|---|
| Issue #471 | Authoritative autonomous objective, stages, safety, human decisions, acceptance, and verification. |
| Draft PR #472 | Parent integration and exact-head human merge gate. |
| Issues #473-#481 | Complete implementation decomposition with dependencies, branches, scopes, TDD, and gates. |
| Parent orchestrator contract | Parent branch/PR ownership, ready queue, worker isolation, stacked PRs, review coverage, and final readiness. |
| Universal GSD/programming loop | Plan-first RED/GREEN/refactor, verification, review, correction, checkpoints, and handoffs. |
| CLI Architecture v2 #397 / PR #438 | First end-to-end consumer/canary; unchanged until its own parent contract authorizes action. |
| #372/#389/#470 and PRs #390/#456 | Explicitly abandoned/superseded historical Go/tmux path; no completion claimed. |
| Legacy shell loops | Temporary rollback path only; deprecation/cutover belongs to #480. |

The replacement covers full autonomous execution, not just read-only validation. Connector-specific,
runtime-service, credential, reverse-ETL, and TUI behavior remains outside Shepherd itself unless a
future parent objective explicitly places it in a bounded child issue.
