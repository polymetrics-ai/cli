# GSD Universal Runtime Loop

Use this workflow to run the same GSD universal programming loop from Claude Code, Codex, OpenCode,
or another agent runtime. Runtime-specific files are activation adapters only; this file owns the
shared workflow policy.

## Runtime Contract

The deterministic GSD scripts are preflight and gate helpers. They do not replace the live
orchestrator.

Every runtime must run this lifecycle:

1. Load `AGENTS.md`, `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`,
   `.planning/STATE.md`, `docs/plans/universal-programming-loop-prd.md`, and
   `docs/prompts/universal-programming-loop-prompts.md`.
2. Run the GSD preflight when available.
   The preflight result is diagnostic only. It does not satisfy orchestration, TDD, verification, or
   completion by itself.
3. Create or update phase `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and
   `RUN-STATE.json`.
4. For behavior changes, capture red test or validation evidence before production edits.
5. For mutating worker roles, create or confirm isolated working directories or git worktrees before
   spawning. A disjoint write scope is not enough if agents share one filesystem checkout.
6. Spawn worker/reviewer subagents for independent tasks when the runtime exposes subagents and
   isolation is available for mutating workers.
7. If subagents are unavailable and mode is `agents`, stop with
   `not_spawned_runtime_capability_missing`.
8. If mutating worker isolation is unavailable and mode is `agents`, stop with
   `not_spawned_runtime_capability_missing` or record `not_spawned_isolation_missing` in the
   parent ledger.
9. If subagents are unavailable or mutating isolation is unavailable and mode is `auto`, run the
   same roles inline or read-only and record fallback.
10. For every cycle after preflight, write one explicit execution decision:
    `spawned`, `read_only_spawned`, `local_critical_path`, or one `not_spawned_*` blocker. Missing
    this decision is a workflow defect.
11. Commit and push coherent green slices after local gates.
12. Stop only for human gates, strict TDD failure, repeated verification failure, or explicit user
   stop.

## Runtime Adapters

### Claude Code

- Slash command: `/gsd:programming-loop`.
- Subagent mechanism: Agent/Task workers.
- Skills: `.claude/skills/*` and compatible project skills.
- Default mode: lean in-session orchestrator delegates heavy work to Agent/Task subagents.
- Universal-loop instruction: read this workflow, the phase artifacts, and the required GSD project
  files before coding; keep the main Claude context as coordinator; use Agent/Task for independent
  implementation, tester, reviewer, security, or reliability roles when scopes are disjoint.
- Parent issue instruction: when a parent issue has ready sub-issues, the coordinator must spawn
  independent workers through Agent/Task or record a `not_spawned_*` blocker.

### Codex

- Project instructions: `AGENTS.md`.
- Custom agents: `.codex/agents/*.toml`.
- Subagent mechanism: Codex subagent tools. Codex only spawns subagents when explicitly asked, so
  parent issue workflows must explicitly say "spawn".
- Skills: Codex agent skills with progressive disclosure.
- Default mode for parent issues: spawn a live parent orchestrator context, then spawn workers for
  every ready subissue with disjoint write scope.
- Universal-loop instruction: load the selected skill bodies before acting, keep `.agents/**` as
  policy source of truth, and call Codex subagent tools explicitly for independent worker/reviewer
  roles instead of only summarizing the plan.
- Parent issue instruction: prompts must explicitly ask Codex to spawn or assign the parent
  orchestrator and every independent ready worker up to `agents.max_threads` or the available
  runtime cap, after the coordinator has created one isolated worktree or working directory per
  mutating worker.

### OpenCode

- Project instructions: `AGENTS.md`.
- Agents: `.opencode/agents/*.md` or `opencode.json`.
- Commands: `.opencode/commands/*.md`; use `subtask: true` only for worker commands that should
  isolate work. Primary orchestrator commands should stay in the main context.
- Skills: `.agents/skills/<name>/SKILL.md`, `.opencode/skills/<name>/SKILL.md`,
  `.claude/skills/<name>/SKILL.md`, or global equivalents.
- Default mode for parent issues: primary orchestrator agent dispatches subagents with bounded
  prompts and records worker handoffs.
- Universal-loop instruction: configure the orchestrator as `mode: primary`, worker/reviewer roles
  as `mode: subagent` or `mode: all`, and allow Task invocation for the worker agents that can own
  disjoint scopes.
- Parent issue instruction: commands that start isolated orchestration work should use
  `subtask: true`; the primary orchestrator still owns the ready queue, spawn decisions, shared
  parent artifacts, and final handoff.
- This repo provides `.opencode/agents/gsd-worker.md` and `.opencode/commands/gsd-worker.md` for
  bounded worker subtasks. The primary orchestrator should invoke that worker command through the
  task/subtask mechanism for independent scopes.

## Active Orchestration Rule

For parent issues with multiple subissues, orchestration is active, not advisory:

- Build ready queue.
- Spawn every independent ready worker up to runtime concurrency limits.
- Treat "spawn" generically: create an Agent/Task worker, Codex subagent job, OpenCode subtask, or
  runtime-native worker context with one issue, one branch, one write scope, one isolated working
  directory, and one handoff template.
- Each orchestration cycle must either spawn/assign at least one ready worker, take the local
  critical-path action that unblocks workers, or record the exact `not_spawned_*` blocker below.
- A long-running worker cannot be kept alive by documentation alone. If the runtime task finishes,
  the orchestrator must either integrate the handoff, spawn the next worker, or record the blocker.
- Keep one parent orchestrator context open until the parent issue reaches human-ready or blocked.
- Do not let completed worker threads remain unclosed after handoff integration.
- If no worker is spawned while work remains, write a blocker with one of:
  `not_spawned_dependency_blocked`, `not_spawned_write_scope_collision`, `not_spawned_human_gate`,
  `not_spawned_isolation_missing`, `not_spawned_runtime_capability_missing`,
  `not_spawned_review_blocked`, or `not_spawned_verification_blocked`.

## Compact Mode

Long-running orchestrators should load the `caveman` skill for agent prose: progress updates,
worker prompts, review summaries, repeated status comments, and handoffs.

Compact mode affects wording and token volume only. It must not change workflow order, verification
requirements, review coverage, or human gates.

Do not use compact mode for exact code, exact commands, exact test output, security warnings,
destructive-action warnings, approval gates, legal/security disclosures, or ordered safety
instructions where shortened wording could change meaning.

## Sources

- Codex skills: https://developers.openai.com/codex/skills
- Codex subagents: https://developers.openai.com/codex/subagents
- Claude Code skills: https://docs.anthropic.com/en/docs/claude-code/skills
- Claude Code subagents: https://docs.anthropic.com/en/docs/claude-code/sub-agents
- OpenCode agents: https://opencode.ai/docs/agents/
- OpenCode commands: https://opencode.ai/docs/commands/
- OpenCode skills: https://opencode.ai/docs/skills/
