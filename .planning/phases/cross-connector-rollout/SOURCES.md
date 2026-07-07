# Source Notes

Accessed: 2026-07-07

## Official Runtime Docs

- Codex skills: https://developers.openai.com/codex/skills
  - Skills use progressive disclosure and load full `SKILL.md` only when selected.
  - Repo skills can live under `.agents/skills`.
- Codex subagents: https://developers.openai.com/codex/subagents
  - Custom agents can live under `.codex/agents/`.
  - Agent settings include concurrency caps such as `agents.max_threads`.
- OpenCode agents: https://opencode.ai/docs/agents/
  - Agents support `primary`, `subagent`, and `all` modes.
  - Task permissions govern which subagents an agent can invoke.
- OpenCode commands: https://opencode.ai/docs/commands/
  - Commands can specify an agent and can force subagent invocation with `subtask: true`.
- OpenCode skills: https://opencode.ai/docs/skills/
  - Skills load on demand through the native skill tool.
  - OpenCode discovers `.agents/skills/<name>/SKILL.md` as a project-compatible skill path.

## Local Source Notes

- `AGENTS.md`: parent issue orchestration must stay active and `gsd-programming-loop` evidence must
  be recorded for behavior-changing work.
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`: active orchestrators
  build ready queues and record spawn decisions.
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`: shared cross-runtime contract.
- `.agents/skills/caveman/SKILL.md`: compact mode preserves exact technical substance and keeps
  warnings/gates clear.
