---
name: polymetrics-issue-delivery
description: Enforce Polymetrics issue-first, bounded dispatch, TDD, exact-head review, and human-gate policy during GSD Pi delivery.
---

# Polymetrics issue delivery

Apply this policy to every GSD planning, execution, validation, UAT, and completion unit.

Before dispatch, read `AGENTS.md`, required skill routing, the issue-agent contract, and the universal
runtime loop. Every subagent dispatch explicitly states: Objective, Output format, Tool guidance,
and Boundaries. Mutating workers require one issue, branch, write scope, and isolated worktree.

For behavior changes, record a failing test or capability probe before the fix, then GREEN and
refactor evidence. Never claim full verification from focused tests. Handoffs are at most 40 lines.

Agents may prepare effect requests but must not directly publish Git/GitHub state. Go Shepherd
grants and emits authorized effects idempotently. No autonomous merge capability exists.

Validation binds the exact candidate head to observed `openai-codex/gpt-5.6-sol` with high reasoning,
local gates, UAT, and milestone validation. A later commit invalidates the attestation. Final
readiness stops at a human merge gate.

Never persist raw prompts, reasoning, credentials, command arguments, or unrestricted tool output.
