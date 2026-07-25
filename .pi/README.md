# Polymetrics Pi Orchestration

This directory configures Pi as a project-local runtime adapter over the shared `.agents/`
agentic delivery system. The `.agents/` directory remains the source of truth for contracts,
workflows, skills, and guardrails.

## Runtime

- Pi CLI: `@earendil-works/pi-coding-agent`
- Project package: `npm:pi-sub-agent@0.1.5`
- Model routing: the orchestrator roles in `.pi/agents/*` carry no `model:` key and inherit the parent session model the driver launches `pi` with (`ORCH_MODEL`); with no `--model`/`ORCH_MODEL` a bare launch falls back to pi's own configured default (historically `openai-codex/gpt-5.5`), which depends on the user's `~/.pi` config / last `/model` selection — `.pi/settings.json` sets no model key
- OpenCode project model: `opencode-go/kimi-k2.7-code`
- OpenCode small model: `opencode-go/deepseek-v4-flash`

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

If you see OpenAI's API quota error, the session is using `OPENAI_API_KEY` instead of the
subscription-backed `openai-codex` provider. Use Pi with `openai-codex/*` models, or unset
`OPENAI_API_KEY` before selecting API-backed `openai/*` models:

```bash
unset OPENAI_API_KEY
pi --provider openai-codex --model gpt-5.5
```

## Usage

Start Pi from the repository root. The orchestration agents request read-only search tools
(`grep`, `find`, `ls`) and the `subagent` extension tool, so enable them explicitly. Pi's default
active tool set is only `read,bash,edit,write`; without the flag below, `pm-scout`/`pm-reviewer`
cannot use `grep`/`find`/`ls`:

```bash
pi --tools read,bash,edit,write,grep,find,ls,subagent --approve
```

Useful prompt templates:

- `/pm-orchestrate`: active parent issue orchestration using project agents.
  Example: `/pm-orchestrate 42` (parent issue number) or `/pm-orchestrate https://github.com/polymetrics-ai/cli/issues/42`.
- `/pm-gsd-loop`: issue-first GSD/TDD implementation loop.
  Example: `/pm-gsd-loop 40` or `/pm-gsd-loop "add DraftIssue to ListProjectItems"`.
- `/pm-review-loop`: Claude/Copilot review disposition loop.
  Example: `/pm-review-loop 74` (PR number) or `/pm-review-loop feat/40-github-projects-discussions`.
- `/pm-auto-loop`: fully automated, resumable delivery loop. The orchestrator model comes from the
  driver, and the workers carry no frontmatter pin — each inherits that same parent session model
  (`ORCH_MODEL`). Given one problem prompt it
  (researches if needed →) plans the parent + sub-issues, creates issues, opens the parent draft PR
  and per-slice sub-PRs, then runs plan→execute→verify→review→correct per task until integrated.
  Handles any task (`problem_type: implementation | connector`).
  Example: `/pm-auto-loop "add a --since flag to pm connectors list"`.
  See `.agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md`.
- `/pm-connector-loop`: connector specialization of the auto-loop — always runs the RESEARCH stage
  (full provider API surface via `pm-web-researcher`) and the 7-slice all-ops CLI-parity decomposition.
  Example: `/pm-connector-loop "twenty (Twenty CRM, https://twenty.com, GraphQL + REST)"`.
  Export `SEARXNG_BASE` (and its token if proxied) first so research can query the audited `searxng`
  connector via `pm`.

### Autonomous driver (unattended, resumable)

Run the loop unattended until it reaches a human gate; it resumes after any interruption
(including token exhaustion) by reconciling from durable state each turn:

```bash
# general task
scripts/pi-auto-loop.sh "add a --since flag to pm connectors list"
# connector (research + all-ops); set the connector loop + SearXNG first
export SEARXNG_BASE=https://<your-searxng-host>
LOOP_CMD=/pm-connector-loop scripts/pi-auto-loop.sh "twenty (Twenty CRM, https://twenty.com, GraphQL + REST)"
scripts/pi-auto-loop.sh --resume        # continue the current run after a stop
```

### Claude-orchestrated driver + Shepherd validator (recommended)

Uses the first-party **Claude Code CLI** (`claude -p`, subscription-backed — no third-party "extra
usage" gate) as the orchestrator and dispatches implementation to a fresh `pi` worker
(`pi --model openai-codex/gpt-5.5`), with an independent **Shepherd supervisor** scoring every step
(revert + replay on a bad step). That `openai-codex` implementation route is currently exhausted;
repointing it with a billing-compliant Claude route is a deliberate separate follow-up (see
`.agents/agentic-delivery/prompts/claude-orchestrator.md`). See
`.agents/agentic-delivery/workflows/shepherd-validator.md`.

```bash
export SEARXNG_BASE=http://localhost:8888
# headless autonomy needs a permission posture; safest is acceptEdits, fully unattended is skip:
export CLAUDE_ARGS="--output-format text --permission-mode acceptEdits"
scripts/claude-auto-loop.sh "twenty (Twenty CRM, https://twenty.com, GraphQL + REST) — full all-ops CLI parity"
scripts/claude-auto-loop.sh --resume
```

The validator writes `.planning/auto-loop/VALIDATION.jsonl` (per-step scores) and
`VALIDATOR-VERDICT.json` (the driver acts on `PROCEED`/`RETRY`/`REVERT`/`HALT`). Requires the local
`claude` CLI to be logged in (`claude -p "ok"` should work).

Model routing is NOT per-agent: `.pi/agents/*.md` carry no `model:` key, so each role inherits the
parent orchestrator session model the driver launches `pi` with (`ORCH_MODEL`), preserving each
agent's declared thinking level. The Shepherd validator turn still defaults to
`openai-codex/gpt-5.6-sol --thinking high` via `VALIDATOR_ARGS`. Choose the fleet's provider by
setting the driver environment (`ORCH_MODEL` / `VALIDATOR_ARGS`) — for example `scripts/pi-auto-loop.sh`
already defaults `ORCH_MODEL` to `anthropic/claude-opus-4-8` — not by re-adding frontmatter pins.
Confirm the exact model IDs your subscription exposes with `/model`. Connector research uses the
repo's `searxng` connector through `pm` (audited path); export `SEARXNG_BASE` (+ token if proxied)
before launching.

### Non-interactive (CI / parent-PR review coverage)

For automated runs, install the subagent tool once (`pi install npm:pi-sub-agent`) and launch with
`--approve` so project-local files are trusted:

```bash
pi -p "Run the GSD verify cycle for phase github-projects-discussions" \
  --tools read,bash,edit,write,grep,find,ls,subagent --approve
```

Project agents under `.pi/agents/` are included by passing `agentScope: "project"` or `"both"` in
the `subagent` tool call (with `confirmProjectAgents: false` for non-interactive runs) — these are
tool parameters, not CLI flags; pi 0.80.x has no `--agentScope`/`--confirmProjectAgents` options.
Only run non-interactively after reviewing and trusting the agents under `.pi/agents/`.

### Compaction and retry

`.pi/settings.json` enables compaction (`reserveTokens: 16384`, `keepRecentTokens: 20000`) for
long-running xhigh orchestration loops, and `maxRetries: 3` for transient provider errors.
Provider-level retry is disabled (`maxRetries: 0`) so Pi-level retry owns backoff.

The first Pi run in this repository asks for project trust because `.pi/settings.json` and
project-local agents can influence runtime behavior. Trust only after reviewing this directory.
