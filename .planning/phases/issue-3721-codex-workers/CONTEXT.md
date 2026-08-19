# CONTEXT — issue #3721 Codex project-local workers

Parent: #3714. Dependency: #3718 / canonical delivery contract. Branch:
`fm/cli-agents-wave-codex-r1`, based on
`refactor/3714-canonical-delivery-flow`.

## Locked decisions

- Generate only the two registered Codex projections:
  `.codex/agents/pm-delivery-worker.toml` and
  `.codex/agents/pm-connector-worker.toml`.
- The canonical JSON contract remains the only policy source. The generator, not a hand-authored
  TOML file, owns their contents and the drift check verifies them.
- Each output is one standalone TOML custom-agent configuration with `name`, `description`, and
  `developer_instructions`. It must set `agents.enabled = false`.
- `agents.enabled = false` is the official documented switch that disables multi-agent tools. This
  proves the generated worker cannot delegate to built-in or ambient custom agents through Codex's
  multi-agent capability; it does not prove that inherited parent MCP/skills configuration is
  absent.
- Use unique `pm-*` names. The official docs document overriding a built-in name but do not state
  a collision rule for identical project and user standalone filenames. The implementation must
  neither depend on nor claim such precedence.
- Codex only loads project `.codex/` layers after the project is trusted. An untrusted project
  therefore fails closed for these workers: users must trust this repository in Codex before
  selecting either role. Do not invent a CLI trust command not documented by the official source.
- Do not touch `~/.codex` or any global config. The legacy project role definitions are deferred
  by this task's Wave 6 scope fence and remain untouched here.

## Official evidence

- [Codex custom agents](https://learn.chatgpt.com/docs/agent-configuration/subagents): standalone
  project TOML, mandatory fields, built-ins, inheritance, and `agents.enabled`.
- [Codex configuration basics](https://developers.openai.com/codex/config-basic/): trusted
  project layers, configuration precedence, and untrusted-project behavior.

## GSD execution note

`scripts/gsd doctor`, `list`, and `sources` resolved the installed lifecycle. Generated
`discuss-phase --auto`, `plan-phase --tdd --skip-research`, and `execute-phase --interactive`
prompts are being executed inline: the canonical contract forbids spawning an orchestrator,
planner, reviewer, verifier, GSD role, or extra worker for this job.
