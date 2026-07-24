# Pi Active Orchestration Loop

Use this workflow when Pi owns a parent issue with subissues.

## Why This Exists

Pi exposes subagents here through the project-local extension configured in `.pi/settings.json`:
`.pi/extensions/pi-sub-agent`. That vendored MIT-licensed copy is derived from
`pi-sub-agent@0.1.5` and carries a local child-session recording modification. It is loaded directly
from the repository after project trust; this route does not install or add the npm package at
runtime. Project agents live under `.pi/agents/`. This file is the Pi adapter for the runtime-generic contract in
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
   - `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
   - `.agents/agentic-delivery/workflows/shepherd-validator.md`
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
   read-only reconnaissance, `pm-verifier` for exact-head verification, and a fresh-context
   `pm-reviewer` for exact-base/head local Codex review. The coordinator
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
   - exact-head verification/review/Shepherd preparation
   - state ledger updates
   - merge eligibility checks
9. When workers finish, integrate handoffs, verify scope, and close completed agent threads.
10. After verification passes, compile a ready v4 exact-base/head/tree manifest, render every
    bounded packet with `scripts/pm-review-system.py render`, pass each rendered stdout unchanged to
    a fresh-context `pm-reviewer`, and run one authenticated v4 synthesis. Disposition every finding
    and repeat the entire exact-head route after any change.
11. Only after the exact-head synthesis is `clean`, run independent Shepherd validation; integrate
    eligible sub-PRs only after `PROCEED` for the same identities.
12. Repeat until all subissues are complete, deferred with issue links, or blocked.
13. Mark the parent PR human-ready only after final exact-head verification, local Codex review,
    Shepherd validation, CI, and remaining human-gate review.

## Canonical PM fallback

Run `scripts/gsd doctor`, `scripts/gsd list`, and source discovery before implementation. If the
registry lacks `programming-loop`, do not invoke or invent it and do not stop at an advisory manual
fallback. `/pm-orchestrate` remains the active owner and executes PLAN → RED → GREEN → REFACTOR →
VERIFY → REVIEW → INTEGRATE with durable state and the same gates.

Claude and GitHub Copilot are not required, requested, or fallback review coverage for this route.
Use `local-codex-review-loop.md` followed by `shepherd-validator.md`.

## Pi Runtime Constraints

- Parallel mode: maximum **8 tasks total**, with up to **4 running concurrently** per `subagent`
  call.
- Chain mode: maximum **8 sequential steps**.
- Recursive `subagent` calls are blocked; the orchestrator is the only spawner.
- `pm-scout` requests only `read`, `grep`, `find`, and `ls`. Exact-head `pm-reviewer` and
  `pm-verifier` also request restricted read-only `bash` for local identity, diff/history, and
  verification commands. The parent Pi session must enable those tools explicitly, e.g.
  `pi --tools read,bash,edit,write,grep,find,ls,subagent --approve`. A subagent can only use tools
  that are active in the parent session.
- Mutating workers must not share the coordinator checkout; use per-worker `cwd` or worktrees.

## Required Spawn Decision

At every parent orchestration turn, write one of:

- `spawned`: mutating workers were spawned, list issue numbers, agent ids, and isolated directories.
- `read_only_spawned`: actual read-only scout/reviewer/verifier contexts were spawned; list their ids and scope.
- `local_critical_path`: the coordinator performed the non-delegable shared or coupled action inline; name it.
- `not_spawned_dependency_blocked`: dependencies block all ready work.
- `not_spawned_write_scope_collision`: ready work would collide.
- `not_spawned_human_gate`: human approval needed.
- `not_spawned_isolation_missing`: safe worker worktree or working directory is unavailable.
- `not_spawned_runtime_capability_missing`: subagent tooling unavailable.
- `not_spawned_review_blocked`: exact-head local Codex, Shepherd, or human gate blocks integration.
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
- `.pi/extensions/pi-sub-agent/README.md` documents the vendored `subagent` tool schema, agent file
  format, concurrency caps, project-agent confirmation behavior, and parent tool-allowlist inheritance.
