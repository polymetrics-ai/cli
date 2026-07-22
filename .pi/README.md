# Polymetrics Pi Orchestration

This directory configures Pi as a project-local runtime adapter over the shared `.agents/`
agentic delivery system. The `.agents/` directory remains the source of truth for contracts,
workflows, skills, and guardrails.

## Runtime

- Pi CLI: `@earendil-works/pi-coding-agent`
- Project package: `npm:pi-sub-agent@0.1.5`
- Default planning/review model: `openai-codex/gpt-5.6-sol` with `xhigh` thinking
- Implementation/correction model: `openai-codex/gpt-5.6-sol` with `high` thinking
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

### Autonomous in-process Shepherd

`/pm-shepherd` is the authoritative replacement program tracked by parent issue #471 and draft
parent PR #472. It uses Pi 0.80.6 public `createAgentSession` APIs inside the current Pi process;
it does not start another `pi` process, use tmux as transport, or rely on the abandoned standalone
Go Shepherd.

The target release owns the whole issue-to-merge loop: research, parent planning, dependency-linked
sub-issues, parent/sub-branches and PRs, isolated parallel implementation workers, GSD
red-green-refactor, verification, review/correction, sub-PR integration, crash reconciliation, and
an exact-head human decision before the parent merge. Implementation roles use
`openai-codex/gpt-5.6-sol`/`high`; planning, research, verification, review, and orchestration roles
use the same model/`xhigh`.

Current branch status: #479 now provides the first bounded autonomous MVP. It schedules disjoint
children at concurrency two, honors dependencies and write-scope collisions, runs implementation
(`high`), verification/review (`xhigh`) as embedded Pi AgentSessions, persists schema-v2 status,
joins stop, resumes unfinished children, and ends at a durable local parent-merge wait. The hardened
schema-v1 read-only canary remains available separately.

Operator flow:

~~~text
/pm-shepherd start --issue 471
/pm-shepherd status --issue 471
/pm-shepherd stop --issue 471
/pm-shepherd resume --issue 471
/pm-shepherd canary --issue 471 --pr 472 --read-only --backend sdk-inproc --experimental
~~~

The MVP intake reads `.planning/shepherd/issue-<N>.json` from the current repository. Example:

~~~json
{
  "schemaVersion": 1,
  "parentIssue": 471,
  "planId": "cli-shepherd-mvp-1",
  "children": [
    {
      "id": "help",
      "issue": 501,
      "title": "Improve help",
      "task": "Add the accepted contextual help behavior and its tests.",
      "dependsOn": [],
      "access": "mutating",
      "writeScopes": ["cmd/help"]
    },
    {
      "id": "tui",
      "issue": 502,
      "title": "Improve TUI",
      "task": "Implement the accepted TUI slice and its tests.",
      "dependsOn": [],
      "access": "mutating",
      "writeScopes": ["internal/tui"]
    }
  ]
}
~~~

This first cut intentionally records an integration checkpoint and a local human wait; it does not
yet publish GitHub issues/PRs, integrate child branches, post the human-gate comment, or observe the
human merge. Those production adapters remain the next bounded slice. There is no parent-to-main
merge capability in the controller.

Mutating workers receive one issue, one branch, one isolated worktree, one declared write scope,
and bounded workspace/host capabilities. Raw prompts, reasoning, credentials, and unrestricted tool
output are never persisted. GitHub authentication remains in the host environment/keychain and is
used only by typed host actions; a token is never passed into a worker prompt.

When a genuine human gate is reached, Shepherd posts one idempotent request on the parent issue
(requirements/scope/authority) or relevant PR (head-specific review/merge) and waits for an
allowlisted response:

~~~text
/shepherd decide <request-id> <option>
~~~

The request is bound to its issue/PR, run generation, allowed options, actor, and exact head when
applicable. Silence, emoji, CI success, review prose, or an agent score is not approval. A parent PR
may merge only after a fresh `approve-merge` decision for its exact verified head and immediate
gate revalidation.

Embedded sessions share the parent Pi process, event loop, heap, environment, and crash domain.
Durability comes from bounded persisted intent plus reconciliation with Git/GitHub truth after
restart, not from process isolation. On macOS, the state root is trusted same-user local state; the
implementation must not claim protection against a hostile same-UID process without native
descriptor-relative filesystem operations.

Registration can be checked without a model, auth, or network call:

~~~bash
printf '{"id":"commands","type":"get_commands"}\n' |
  PI_OFFLINE=1 pi --mode rpc --no-session --approve \
    --no-extensions --no-skills --no-prompt-templates --no-context-files \
    -e .pi/extensions/shepherd/index.ts |
  jq -e 'select(.id=="commands") |
    any(.data.commands[]; .name=="pm-shepherd" and .source=="extension")'
~~~

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
usage" gate) as the orchestrator and **Codex** (`pi --model openai-codex/gpt-5.6-sol --thinking high`) for
implementation, with an independent **Shepherd supervisor** scoring every step (revert + replay on a
bad step). See `.agents/agentic-delivery/workflows/shepherd-validator.md`.

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

Model routing is per-agent in `.pi/agents/*.md`; the Codex-only Shepherd profile pins project
agents to `openai-codex/gpt-5.6-sol` with the exact role thinking level, while the legacy Shepherd
validator defaults to `openai-codex/gpt-5.6-sol --thinking high`. Confirm the exact model IDs your
subscription exposes with `/model`, then set them once in the agent frontmatter and in the driver
environment (`ORCH_MODEL` / `VALIDATOR_ARGS`). Connector research uses the repo's `searxng`
connector through `pm` (audited path); export `SEARXNG_BASE` (+ token if proxied) before launching.

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
