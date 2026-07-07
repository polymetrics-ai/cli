# Codex Active Orchestration Loop

Use this workflow when Codex owns a parent issue with subissues.

## Why This Exists

Codex subagents are available, but Codex only spawns them when explicitly asked. A repo-level parent
orchestrator contract is therefore not enough. The coordinator must create and steer an actual
orchestrator context and workers. This file is a Codex adapter for the runtime-generic contract in
`.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Loop

1. Confirm parent issue, parent branch, and parent PR.
2. Open or reuse a live parent orchestrator context.
3. Load:
   - `AGENTS.md`
   - `.agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml`
   - `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
   - `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
   - `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
   - `.agents/skills/caveman/SKILL.md`
4. Build ready queue from parent issue and orchestration state.
5. Create or confirm an isolated working directory for every worker that may edit files. For Codex,
   prefer a git worktree per worker branch, for example
   `git worktree add ../<repo>-worktrees/<issue>-<slug> -b <branch> <parent-branch>`.
6. Spawn one worker per independent ready subissue, up to the available concurrency cap.
7. Assign each worker one issue, one branch, one write scope, one working directory, and one handoff
   template.
8. While workers run, the main coordinator does non-overlapping work only:
   - parent PR status
   - review routing
   - state ledger updates
   - merge eligibility checks
9. When workers finish, integrate handoffs, verify scope, and close completed agent threads.
10. Merge eligible sub-PRs into parent branch.
11. Wait for parent PR automatic review coverage for integrated ranges when required.
12. Repeat until all subissues are complete, deferred with issue links, or blocked.
13. Mark parent PR human-ready only after final verification and automated review disposition.

## Required Spawn Decision

At every parent orchestration turn, write one of:

- `spawned`: workers were spawned, list issue numbers and agent ids.
- `not_spawned_dependency_blocked`: dependencies block all ready work.
- `not_spawned_write_scope_collision`: ready work would collide.
- `not_spawned_human_gate`: human approval needed.
- `not_spawned_isolation_missing`: safe worker worktree or working directory is unavailable.
- `not_spawned_runtime_capability_missing`: subagent tooling unavailable.
- `not_spawned_review_blocked`: CodeRabbit/Copilot/human review route blocks integration.
- `not_spawned_verification_blocked`: checks or local gates block integration.

Missing this decision is a workflow defect.

## Codex Adapter

If a named custom agent is exposed, use it. If not, spawn `default` with the parent orchestrator YAML
and this workflow pasted as the task contract.

Do not give a code-writing worker the coordinator's current repository path. Codex subagents can
share the same filesystem context unless the prompt and setup give them a separate worktree or
explicitly read-only scope. If no isolated worktree is available, do not spawn mutating workers;
record `not_spawned_isolation_missing` and either run the slice locally or use read-only explorer
agents.

Prefer compact prompts:

```text
Use caveman compact mode for status and handoff. Preserve exact commands, paths, tests, safety gates.
```

Compact prompts must not shorten exact code, exact commands, exact test output, security warnings,
destructive-action warnings, approval gates, or ordered safety instructions.

## Source Notes

- Codex skills use progressive disclosure and load full `SKILL.md` instructions only when selected:
  https://developers.openai.com/codex/skills
- Codex custom agents can be project-scoped under `.codex/agents/`, and Codex agent settings expose
  concurrency caps such as `agents.max_threads`: https://developers.openai.com/codex/subagents
