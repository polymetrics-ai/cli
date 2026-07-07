# GSD Universal Runtime Loop

Use this workflow to run the same GSD universal programming loop from Claude Code, Codex, OpenCode,
or another agent runtime.

## Runtime Contract

The deterministic GSD scripts are preflight and gate helpers. They do not replace the live
orchestrator.

Every runtime must run this lifecycle:

1. Load `AGENTS.md`, `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`,
   `.planning/STATE.md`, `docs/plans/universal-programming-loop-prd.md`, and
   `docs/prompts/universal-programming-loop-prompts.md`.
2. Run the GSD preflight when available.
3. Create or update phase `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and
   `RUN-STATE.json`.
4. For behavior changes, capture red test or validation evidence before production edits.
5. Spawn worker/reviewer subagents for independent tasks when the runtime exposes subagents.
6. If subagents are unavailable and mode is `agents`, stop with `failed_runtime_capability`.
7. If subagents are unavailable and mode is `auto`, run the same roles inline and record fallback.
8. Commit and push coherent green slices after local gates.
9. Stop only for human gates, strict TDD failure, repeated verification failure, or explicit user
   stop.

## Runtime Adapters

### Claude Code

- Slash command: `/gsd:programming-loop`.
- Subagent mechanism: `Task`.
- Skills: `.claude/skills/*` and compatible project skills.
- Default mode: lean in-session orchestrator delegates heavy work to Task subagents.

### Codex

- Project instructions: `AGENTS.md`.
- Custom agents: `.codex/agents/*.toml`.
- Subagent mechanism: Codex subagent tools. Codex only spawns subagents when explicitly asked, so
  parent issue workflows must explicitly say "spawn".
- Skills: Codex agent skills with progressive disclosure.
- Default mode for parent issues: spawn a live parent orchestrator context, then spawn workers for
  every ready subissue with disjoint write scope.

### OpenCode

- Project instructions: `AGENTS.md`.
- Agents: `.opencode/agents/*.md` or `opencode.json`.
- Commands: `.opencode/commands/*.md`; use `subtask: true` for commands that should isolate work.
- Skills: `.agents/skills/<name>/SKILL.md`, `.opencode/skills/<name>/SKILL.md`,
  `.claude/skills/<name>/SKILL.md`, or global equivalents.
- Default mode for parent issues: primary orchestrator agent dispatches subagents with bounded
  prompts and records worker handoffs.

## Active Orchestration Rule

For parent issues with multiple subissues, orchestration is active, not advisory:

- Build ready queue.
- Spawn every independent ready worker up to runtime concurrency limits.
- Keep one parent orchestrator context open until the parent issue reaches human-ready or blocked.
- Do not let completed worker threads remain unclosed after handoff integration.
- If no worker is spawned while work remains, write a blocker with one of:
  `dependency_blocked`, `write_scope_collision`, `human_gate`, `runtime_capability_missing`,
  `review_blocked`, or `verification_blocked`.

## Compact Mode

Long-running orchestrators should load the `caveman` skill for progress updates, worker prompts,
and handoffs. Compact mode is forbidden when it would make safety gates, destructive actions,
security warnings, or ordered instructions ambiguous.

## Sources

- Codex skills: https://developers.openai.com/codex/skills
- Codex subagents: https://developers.openai.com/codex/subagents
- Claude Code skills: https://docs.anthropic.com/en/docs/claude-code/skills
- Claude Code subagents: https://docs.anthropic.com/en/docs/claude-code/sub-agents
- OpenCode agents: https://opencode.ai/docs/agents/
- OpenCode commands: https://opencode.ai/docs/commands/
- OpenCode skills: https://opencode.ai/docs/skills/
