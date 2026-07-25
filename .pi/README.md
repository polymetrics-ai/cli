# Polymetrics Pi Orchestration

This directory configures Pi as a project-local runtime adapter over the shared `.agents/`
agentic delivery system. The `.agents/` directory remains the source of truth for contracts,
workflows, skills, and guardrails.

## Runtime

- Pi CLI: `@earendil-works/pi-coding-agent`
- Project subagent extension: vendored `.pi/extensions/pi-sub-agent`, derived from the
  MIT-licensed `pi-sub-agent@0.1.5` package with a local child-session recording modification;
  `.pi/settings.json` loads this repository path directly
- Default Pi coordinator model: `openai-codex/gpt-5.6-sol` with `xhigh` thinking
- Implementation agents: `openai-codex/gpt-5.6-sol` with `high` thinking
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
pi --provider openai-codex --model gpt-5.6-sol --thinking xhigh
```

## Usage

Start Pi from the repository root. The orchestration agents request read-only search tools
(`grep`, `find`, `ls`) and the `subagent` extension tool, so enable them explicitly. Pi's default
active tool set is only `read,bash,edit,write`; without the flag below, `pm-scout`, `pm-reviewer`,
and `pm-verifier` cannot use `grep`/`find`/`ls` (reviewer/verifier also retain restricted read-only
`bash` for local identity, history/diff, and verification commands):

```bash
pi --model openai-codex/gpt-5.6-sol --thinking xhigh \
  --tools read,bash,edit,write,grep,find,ls,subagent --approve
```

Useful prompt templates:

- `/pm-orchestrate`: active parent issue orchestration using project agents.
  Example: `/pm-orchestrate 42` (parent issue number) or `/pm-orchestrate https://github.com/polymetrics-ai/cli/issues/42`.
- `/pm-gsd-loop`: issue-first GSD/TDD implementation loop.
  Example: `/pm-gsd-loop 40` or `/pm-gsd-loop "add DraftIssue to ListProjectItems"`.
- `/pm-review-loop`: v4 compile plus exact packet render, fresh-context local Codex responses, and
  one authenticated synthesis; only a clean exact head proceeds to independent Shepherd validation.
  Example: `/pm-review-loop 74 <exact-base-sha> <exact-head-sha>`.
- `/pm-auto-loop`: fully automated, resumable delivery loop. The orchestrator model comes from the
  driver, and worker models come from `.pi/agents/*.md` frontmatter. Given one problem prompt it
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

### Canonical PM driver + Shepherd validator

Use this for unattended, independently validated Pi orchestration. The driver advances one durable
stage per turn, runs exact-head verification and fresh-context local Codex review, runs a separate
Sol/xhigh Shepherd validator after every turn, replays rejected transitions, and stops at the final
human gate. Implementation subagents use Sol/high from project frontmatter.

```bash
scripts/pi-shepherd-loop.sh "<problem or existing parent issue continuation>"
scripts/pi-shepherd-loop.sh --resume
```

`--resume` requires an existing run's `.planning/auto-loop/PROMPT.txt`. Environment overrides such
as `AUTO_LOOP_STATE_DIR`, `ORCH_MODEL`, `ORCH_THINKING`, and `VALIDATOR_ARGS` must be repeated on
resume because the prompt file does not persist them.

### Legacy driver migration

`scripts/claude-auto-loop.sh` remains only for historical trace replay. It is not a current or
forward PM orchestration or review route. Do not request Claude or GitHub Copilot as required,
fallback, or substitute PM coverage. Use `scripts/pi-shepherd-loop.sh`,
`.agents/agentic-delivery/workflows/local-codex-review-loop.md`, and
`.agents/agentic-delivery/workflows/shepherd-validator.md` instead.

The validator writes `.planning/auto-loop/VALIDATION.jsonl` (per-step scores) and
`VALIDATOR-VERDICT.json` (the driver acts on `PROCEED`/`RETRY`/`REVERT`/`HALT`).

Run `scripts/gsd doctor` and `scripts/gsd list` before implementation. If `programming-loop` is
absent, `/pm-orchestrate` owns PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE; never
invent the unavailable command.

Model routing is per-agent in `.pi/agents/*.md`; every Codex-only Shepherd role uses
`openai-codex/gpt-5.6-sol`. Implementation agents use `high`; the coordinator, Shepherd validator,
planning, research, verification, and review agents use `xhigh`. Confirm availability with
`pi --list-models gpt-5.6-sol`; driver overrides are `ORCH_MODEL`, `ORCH_THINKING`, and
`VALIDATOR_ARGS`. Connector research uses the repo's `searxng`
connector through `pm` (audited path); export `SEARXNG_BASE` (+ token if proxied) before launching.

### Non-interactive (CI / parent-PR review coverage)

For automated runs, review and trust the project-local agents and the vendored
`.pi/extensions/pi-sub-agent` source, then launch with `--approve`. `.pi/settings.json` loads the
local extension; do **not** run `pi install npm:pi-sub-agent` for this project, because that adds a
second package route, can duplicate the tool, omits the local modification, and crosses the new-dependency gate:

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
