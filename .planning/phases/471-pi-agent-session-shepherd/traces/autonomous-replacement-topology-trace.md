# Autonomous Replacement Topology Trace

Date: 2026-07-21
Actor: parent orchestrator

## User decision

The Go Shepherd is abandoned, not a durable backend for a read-only Pi companion. Build a complete
autonomous replacement with in-process Pi `AgentSession` workers, dependency-aware parallelism,
GSD/TDD/review/correction, durable issue/PR human decisions, and parent delivery through `main`
only after the human merge gate.

## GitHub mutations

- Rewrote #471 as the authoritative parent.
- Opened draft parent PR #472 from `feat/471-pi-agent-session-shepherd` to `main`.
- Created child issues #473-#481 with explicit dependencies, branches, PR base, scopes, TDD,
  verification, skills, and gates.
- Closed #470, #389, and #372 as `not_planned` with supersession comments.
- Closed draft PRs #456 and #390 unmerged with supersession comments.

No completion was claimed for the abandoned program. No branch, worktree, runtime state, or
historical commit was deleted. PR #391 remains immutable merged history on the abandoned parent
branch. Legacy shell/GSD records and #397/#438 were not closed or rewritten.

## Ready queue

#473 is the only ready issue. #474-#477 become a four-lane parallel wave after #473 integration;
#478, #479, #480, and #481 follow their recorded dependencies. The existing uncommitted foundation
remediation is assigned to #473.
