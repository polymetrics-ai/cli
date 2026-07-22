# Cross-Connector Rollout Plan

Issue: #42
Parent issue: #44
Parent branch: `feat/44-github-cli-parity`
Worker branch: `docs/42-harden-universal-orchestration`
Task type: docs-only workflow hardening

## Objective

Make the agentic delivery workflow runtime-generic across Codex, OpenCode, Claude Code, and future
runtimes while preserving active parent orchestration and safe compact handoffs.

## Scope

- Clarify shared GSD runtime adapter instructions in `.agents/agentic-delivery/**`.
- Clarify parent orchestrator duty to spawn every independent ready worker up to runtime limits.
- Clarify that mutating workers need isolated worktrees or working directories, not only disjoint
  file scopes.
- Clarify `caveman` compact mode as prose/handoff compression only.
- Record local and official source notes for Codex and OpenCode runtime features.
- Keep the change generic, not specific to one connector or one provider.

## Non-Goals

- No production connector definitions.
- No Go, website, runtime service, or generated-file edits.
- No manual Claude review trigger.
- No changes to `.codex/**` or `.opencode/**` adapter files in this slice.

## Docs-Only Loop

1. Plan and ledger before doc edits.
2. Patch shared contracts/workflows and compact-mode guidance.
3. Validate Markdown/YAML/JSON syntax for changed files.
4. Commit, push, open a stacked PR to `feat/44-github-cli-parity`.
5. Wait for automatic checks/review when available; do not manually trigger Claude.

## Acceptance Criteria

- Runtime adapters in the shared workflow name universal programming-loop instructions for Claude
  Code, Codex, and OpenCode.
- Parent orchestration is active: orchestrator builds the queue and spawns/assigns independent
  ready workers, or records a `not_spawned_*` blocker.
- Mutating workers are not spawned into the coordinator checkout; if isolation is unavailable, the
  orchestrator records `not_spawned_isolation_missing`.
- Caveman guidance says it compresses agent prose, prompts, status, and handoffs only.
- Caveman guidance forbids compacting exact code, exact commands, exact test output, ordered safety
  gates, security warnings, destructive-action warnings, or approval gates.
- Source notes use official Codex and OpenCode docs URLs only for current runtime feature claims.
