# Pi Active Orchestration Loop

Use this workflow when Pi owns a parent issue with subissues.

## Why This Exists

Pi ships with subagents only when a package such as `pi-sub-agent` is enabled. The project runtime
adds `npm:pi-sub-agent@0.1.5` and project agents under `.pi/agents/`. This file is the Pi adapter
for the runtime-generic contract in
`.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and the parent orchestrator
contract in `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`.

## Loop

1. Confirm parent issue, parent branch, and parent PR.
2. Open or reuse a live parent orchestrator context in the main Pi session.
3. Load:
   - `AGENTS.md`
   - `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
   - `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
   - `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
   - `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
   - `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
   - `.agents/agentic-delivery/workflows/claude-review-loop.md`
   - `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md`
   - `.agents/skills/caveman/SKILL.md`
4. Build ready queue from parent issue and orchestration state.
5. Create or confirm an isolated working directory for every worker that may edit files. For Pi,
   pass a per-task `cwd` to the `subagent` tool. Prefer a git worktree per worker branch, for
   example `git worktree add ../<repo>-worktrees/<issue>-<slug> -b <branch> <parent-branch>`.
   `.planning/config.json` sets `use_worktrees: false`, so the project's default isolation is a
   per-task `cwd` (a sibling checkout), not a git worktree; either satisfies the contract. A
   mutating worker without its own `cwd` must be recorded as `not_spawned_isolation_missing`.
6. Spawn one worker per independent ready subissue through the `subagent` tool, up to the Pi
   runtime limits. Dispatch `pm-gsd-worker` for mutating implementation, `pm-scout` for
   read-only reconnaissance, and `pm-reviewer` for read-only adversarial review. The coordinator
   must call `subagent` in the current turn; writing a plan that says a worker should exist is
   not a spawn decision. Inline role passes are `local_critical_path` or
   `not_spawned_runtime_capability_missing`, never `spawned`.
7. Assign each worker one issue, one branch, one write scope, one working directory, and one
   handoff template. Use `agentScope: "both"` (or `"project"`) so project agents from `.pi/agents/`
   are visible. In non-interactive runs, project agents are blocked unless
   `confirmProjectAgents: false` is set; only set it after reviewing and trusting the project
   agents.
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

## Pi Runtime Constraints

- Parallel mode: maximum **8 tasks total**, with up to **4 running concurrently** per `subagent`
  call.
- Chain mode: maximum **8 sequential steps**.
- Recursive `subagent` calls are blocked; the orchestrator is the only spawner.
- Read-only agents (`pm-scout`, `pm-reviewer`) request `grep`, `find`, and `ls`. The parent Pi
  session must enable those tools explicitly, e.g.
  `pi --tools read,bash,edit,write,grep,find,ls,subagent --approve`. A subagent can only use tools
  that are active in the parent session.
- Mutating workers must not share the coordinator checkout; use per-worker `cwd` or worktrees.

## Required Spawn Decision

At every parent orchestration turn, write one of:

- `spawned`: workers were spawned, list issue numbers and agent ids.
- `not_spawned_dependency_blocked`: dependencies block all ready work.
- `not_spawned_write_scope_collision`: ready work would collide.
- `not_spawned_human_gate`: human approval needed.
- `not_spawned_isolation_missing`: safe worker worktree or working directory is unavailable.
- `not_spawned_runtime_capability_missing`: subagent tooling unavailable.
- `not_spawned_review_blocked`: Claude/Copilot/human review route blocks integration.
- `not_spawned_verification_blocked`: checks or local gates block integration.

Missing this decision is a workflow defect. The decision must include evidence: agent id(s),
worker directory, issue number, write scope, or the exact blocker reason.

## Compact Mode

Use caveman compact mode for status and handoff. Preserve exact commands, paths, tests, safety
gates, security warnings, destructive-action warnings, and approval gates.

## Source Notes

- Pi prompt templates: `docs/prompt-templates.md`
- Pi settings: `docs/settings.md`
- Pi packages: `docs/packages.md`
- `pi-sub-agent` README documents the `subagent` tool schema, agent file format, concurrency caps,
  project-agent confirmation behavior, and parent tool-allowlist inheritance.
