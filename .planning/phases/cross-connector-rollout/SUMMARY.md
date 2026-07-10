# Summary

Status: completed.

## Changes

- Clarified the shared GSD universal runtime loop as the source of truth for Claude Code, Codex,
  OpenCode, and future runtimes.
- Made parent orchestration explicitly runtime-generic: `spawn` can mean Claude Code `Task`, Codex
  subagent job, OpenCode subtask, or a future isolated worker context.
- Added the live Codex lesson: code-writing workers need their own worktree or working directory;
  shared-checkout mutation is a blocker, even when file scopes are disjoint.
- Hardened Caveman compact-mode rules so compacting only affects agent prose/status/handoffs and
  never exact commands, code, tests, review findings, security warnings, destructive-action
  warnings, ordered safety gates, or approval gates.
- Updated the parent orchestrator and cross-connector rollout agent specs with active queue/spawn
  evidence, runtime adapter notes, docs-only GSD evidence, and source-note expectations.
- Added the `cross-connector-rollout` phase ledger for plan, TDD exemption, verification, prompt
  notes, source notes, and run state.

## Verification

- PASS: `git diff --check -- .agents/agentic-delivery .agents/skills/caveman .planning/phases/cross-connector-rollout`
- PASS: YAML parse for `.agents/agentic-delivery` and `.planning/phases/cross-connector-rollout`
- PASS: `jq empty .planning/phases/cross-connector-rollout/RUN-STATE.json`
- PASS: follow-up review after a live Codex worker collision confirmed workflow now requires
  isolated worktrees or records `not_spawned_isolation_missing`.

## Review Route

Stacked sub-PR should target `feat/44-github-cli-parity` and use automatic checks/review only. Do
not manually trigger Claude.
