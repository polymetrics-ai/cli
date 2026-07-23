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

### Bounded workflow-engine development orchestration

`.pi/settings.json` pins `npm:pi-workflow-engine@0.12.0` for project-local bounded analysis,
typed synthesis, and independent review. The inspected npm artifact resolves to
`pi-workflow-engine-0.12.0.tgz` with integrity
`sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==`.
The package requires Pi 0.80.10 or newer; this repository deliberately narrows Shepherd to the
stable interval `>=0.80.10 <0.80.11` until another release is explicitly tested.

This is a partial adoption. Workflow-engine runs are non-authoritative development/review
orchestration. Their `.pi/.workflow-runs` records, journals, run IDs, isolated patches, retries,
background state, and disposable worktrees never replace Shepherd state, external-effect receipts,
recovery, persistent issue worktrees, exact-head evidence, or human gates. Production Shepherd does
not import workflow-engine or its undocumented internals. It retains `ProductionAgentSessionPort`
and uses public Pi APIs directly with claimed cwd, exact scoped host tools, prompt completion,
typed bound handoffs, cancellation, and abort/join-before-release.

### Autonomous in-process Shepherd

`/pm-shepherd` is the authoritative replacement program tracked by parent issue #471 and draft
parent PR #472. It uses bounded-compatible Pi 0.80.10 public `createAgentSession` APIs inside the current Pi process;
it does not start another `pi` process, use tmux as transport, or rely on the abandoned standalone
Go Shepherd.

The production path establishes a reviewed schema-2 plan, creates or reconciles child issues, runs
dependency-ready children in isolated worktrees, executes fixed bounded verification, opens stacked
child PRs, requires a clean independent review of each exact head, applies bounded corrections, and
integrates accepted child PRs into the non-default parent branch. If the plan is absent, a read-only
xhigh planning AgentSession reads bounded authoritative GitHub issue facts and proposes semantic
children without issue numbers or host authority. The host validates that proposal, creates or
reconciles marker-bound child issues, inserts only the returned GitHub issue numbers, and atomically
publishes the ignored plan file. It then verifies and reviews the exact parent head, publishes one
durable human decision request, and waits. It has no operation that merges the parent PR into the
default branch.

Before starting:

- Launch Pi 0.80.10 from the clean canonical Git worktree on the intended non-default parent branch.
  Authenticate Pi's configured models and `gh` in the host environment; never put credentials in a
  plan or prompt. Plan bootstrap requires authoritative `admin` or `maintain` repository permission.
- Ensure the parent GitHub issue and non-default parent branch exist. On `start`, the host creates or
  reconciles the one marker-bound draft parent PR to the authoritative default branch before durable
  controller state is created; ambiguous or conflicting PR evidence fails closed.
- A valid existing `.planning/shepherd/issue-<N>.json` is reused. When it is absent, `start` generates
  it through the bounded planning flow above. An existing invalid or conflicting file is never
  overwritten automatically. `start` refuses an existing run; use `resume` for persisted work.

The command surface is:

~~~text
/pm-shepherd
/pm-shepherd help
/pm-shepherd start --issue 471 --backend sdk-inproc --max-concurrency 2 --timeout-seconds 900
/pm-shepherd status --issue 471
/pm-shepherd stop --issue 471
/pm-shepherd resume --issue 471 --backend sdk-inproc --max-concurrency 2 --timeout-seconds 900
/pm-shepherd canary --issue 471 --pr 472 --read-only --backend sdk-inproc --experimental
~~~

`--max-concurrency` is limited to 1 or 2 and defaults to 2. `--timeout-seconds` is limited to
30–3600 and defaults to 900. Resume must repeat any non-default concurrency and timeout values from
the original start. `start` and `resume` still accept `--pr N` as a compatibility input, but the
production controller does not use it as authority; the plan and authoritative GitHub evidence bind
the parent PR. Only one extension run may be active in the Pi process. Bare/help and `status` are
read-only; `status` renders persisted state and never dispatches an AgentSession or consumes a human
decision. Use `resume` after a human action.

The schema-2 plan has exact fields: unknown or missing fields fail validation. Every top-level child
is mutating, has a non-empty repository-relative write scope and at least one fixed verification
command, and declares its own attempt and correction budgets. This three-child example permits
`state` and `pipeline` to run together; `controller` waits for both:

~~~json
{
  "schemaVersion": 2,
  "planId": "issue-471-production",
  "parentIssue": 471,
  "repository": "owner/repository",
  "title": "Complete the production Shepherd",
  "objective": "Deliver and verify the reviewed issue 471 production plan.",
  "parentBranch": "feat/471-pi-agent-session-shepherd",
  "parentBaseBranch": "main",
  "actorAllowlist": ["maintainer"],
  "decisionExpiresAt": "2027-12-31T23:59:59Z",
  "children": [
    {
      "id": "state",
      "issue": 501,
      "title": "Persist production state",
      "task": "Implement the reviewed production state slice and its tests.",
      "slug": "production-state",
      "dependsOn": [],
      "access": "mutating",
      "writeScopes": [
        ".pi/extensions/shepherd/autonomous-production-state.ts",
        ".pi/extensions/shepherd/autonomous-production-state.test.ts"
      ],
      "requiredSkills": ["architecture-patterns", "javascript-testing-patterns"],
      "verification": [
        {
          "id": "state-tests",
          "executable": "node",
          "args": ["--test", ".pi/extensions/shepherd/autonomous-production-state.test.ts"],
          "cwd": ".",
          "timeoutMs": 30000,
          "maxOutputBytes": 1048576
        }
      ],
      "humanGates": [],
      "maxAttempts": 2,
      "maxCorrections": 1
    },
    {
      "id": "pipeline",
      "issue": 502,
      "title": "Compose the child pipeline",
      "task": "Implement the reviewed child lifecycle and exact-head review tests.",
      "slug": "production-child-pipeline",
      "dependsOn": [],
      "access": "mutating",
      "writeScopes": [
        ".pi/extensions/shepherd/production-child-pipeline.ts",
        ".pi/extensions/shepherd/production-child-pipeline.test.ts"
      ],
      "requiredSkills": ["architecture-patterns", "javascript-testing-patterns"],
      "verification": [
        {
          "id": "pipeline-tests",
          "executable": "node",
          "args": ["--test", ".pi/extensions/shepherd/production-child-pipeline.test.ts"],
          "cwd": ".",
          "timeoutMs": 30000,
          "maxOutputBytes": 1048576
        }
      ],
      "humanGates": [],
      "maxAttempts": 2,
      "maxCorrections": 1
    },
    {
      "id": "controller",
      "issue": 503,
      "title": "Wire the production controller",
      "task": "Integrate the reviewed production state and child lifecycle through the controller.",
      "slug": "production-controller",
      "dependsOn": ["state", "pipeline"],
      "access": "mutating",
      "writeScopes": [
        ".pi/extensions/shepherd/production-controller.ts",
        ".pi/extensions/shepherd/production-controller.test.ts"
      ],
      "requiredSkills": ["architecture-patterns", "javascript-testing-patterns"],
      "verification": [
        {
          "id": "controller-tests",
          "executable": "node",
          "args": ["--test", ".pi/extensions/shepherd/production-controller.test.ts"],
          "cwd": ".",
          "timeoutMs": 30000,
          "maxOutputBytes": 1048576
        }
      ],
      "humanGates": [],
      "maxAttempts": 2,
      "maxCorrections": 1
    }
  ]
}
~~~

The validator permits at most 64 children, 1–10 attempts, and 1–5 corrections. It rejects dependency
cycles, traversal, absolute paths, backslashes, control text, sparse or oversized payloads, and
top-level read-only children. Verification is not an agent-provided shell string: production accepts
only closed Node test-runner, Go test/vet/build, or allowlisted Make quality-gate recipes; resolves
canonical host-owned executables; passes fixed argv without a shell; canonicalizes `cwd` inside the
worktree; sanitizes the environment; caps each command at 120 seconds and 4 MiB; terminates the
POSIX process group on timeout/cancellation; and hard-bounds settlement.

At each scheduling boundary, Shepherd selects a deterministic dependency-ready set subject to the
concurrency cap and canonical write-scope collisions. Disjoint mutating worktrees may coexist; an
overlapping scope waits, and each completed, failed, or stopped child releases only its own lease.
Persisted status explains whether work is idle for capacity, dependencies, or a write-scope
collision. When an integrated sibling advances the parent, stale children refresh/reclaim their
workspace and must repeat verification and exact-head review before integration.

Child integration never calls GitHub's head-only pull-request merge mutation. Shepherd builds one
deterministic two-parent merge from the exact reviewed base/head, rechecks the live remote default,
and advances only the non-default parent ref with an exact `--force-with-lease` old-SHA guard.
GitHub is then used only to observe the exact merge/parent ref and publish the durable receipt. A
restart after the Git ref update but before that receipt reuses the same proven merge; an unrelated
parent advance is preserved and sends the child through refresh, verification, and review again.

Retryable failures consume `maxAttempts`; failed verification or review findings consume
`maxCorrections`, and review findings require recorded dispositions plus a clean review of the
resulting exact head. Exhaustion creates a durable issue-bound child gate instead of treating prose
or partial output as success. An allowlisted human answers the exact, unedited child-issue comment
with one of:

~~~text
/shepherd decide <request-id> authorize-one-retry
/shepherd decide <request-id> abort-child
~~~

`stop` cancels and joins accepted intake, workspace, AgentSession, verification, Git/GitHub,
decision, and backoff work before persisting the resumable boundary. `resume` validates the original
plan digest, repository/branch/scope ownership, policy, generation fences, and unfinished effect
journal before scheduling anything new. Prepared or observed commit, push, PR, integration, and
decision effects are reconciled against authoritative Git/GitHub state so a timeout or restart does
not publish the same effect twice.

After all exact child receipts and parent evidence pass, Shepherd posts one idempotent exact-head
request on the parent PR and waits for an allowlisted response:

~~~text
/shepherd decide <request-id> approve-merge
/shepherd decide <request-id> reject
~~~

The command must be the entire body of one unedited comment from an allowlisted non-bot actor and is
bound to its issue/PR, run generation, allowed options, and exact head. Silence, emoji, CI success,
review prose, duplicate commands, an edited comment, or an agent score is not approval. An
`approve-merge` decision authorizes only the human-owned merge: Shepherd remains waiting until a
human merges that exact parent head through the normal GitHub process. A later
`/pm-shepherd resume --issue N` observes authoritative merge evidence and only then marks the run
complete. A moved head invalidates the gate and requires fresh verification, review, and approval.

Mutating workers receive one issue, one branch, one isolated worktree, one declared write scope,
bounded workspace edit/write tools, and an ID-only `host_verify` capability for RED→GREEN reruns.
The independent verification session receives repository reads plus the same ID-only capability,
but no workspace mutation tool, and must request every immutable verification ID in exact order.
The schema-2 input may contain task text and validated verification argv, so neither may contain
secret material. Persisted production state remains schema version 1
with kind `production_autonomous`; it records bounded summaries, digests, receipts, counters, and
gate bindings, never the plan task text, prompts, reasoning, raw model output, credentials, or
unrestricted tool output. GitHub authentication remains host-only and is never passed into a worker
prompt.

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
