# Polymetrics Pi Orchestration

This directory configures Pi as a project-local runtime adapter over the shared `.agents/`
agentic delivery system. The `.agents/` directory remains the source of truth for contracts,
workflows, skills, and guardrails.

## Runtime

- Pi CLI: `@earendil-works/pi-coding-agent`
- Project package: `npm:pi-sub-agent@0.1.5`
- Default Pi model: `openai/gpt-5.5` with `xhigh` thinking
- OpenCode project model: `openai/gpt-5.5`
- OpenCode small model: `openai/gpt-5.3-codex-spark`

## Authentication

Do not store API keys or OAuth tokens in this repository.

For ChatGPT Plus/Pro Codex subscription auth:

```bash
cd /Users/karthiksivadas/Development/polymetrics-cli-agents/connector-cli-parity-research
pi
/login
```

Select `ChatGPT Plus/Pro (Codex)`. Pi stores OAuth tokens in `~/.pi/agent/auth.json`.

For OpenCode:

```bash
opencode providers login
```

Select OpenAI or OpenCode Go/Zen as needed. OpenCode stores credentials in
`~/.local/share/opencode/auth.json`.

For Pi's OpenCode provider path, set `OPENCODE_API_KEY` in the shell before launching Pi when
using `opencode` or `opencode-go` models. Do not commit this value.

## Usage

Start Pi from the repository root:

```bash
pi --approve
```

Useful prompt templates:

- `/pm-orchestrate`: active parent issue orchestration using project agents.
- `/pm-gsd-loop`: issue-first GSD/TDD implementation loop.
- `/pm-review-loop`: CodeRabbit/Copilot review disposition loop.

The first Pi run in this repository asks for project trust because `.pi/settings.json` and
project-local agents can influence runtime behavior. Trust only after reviewing this directory.
