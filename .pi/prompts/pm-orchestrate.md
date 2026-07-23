---
description: Active parent issue orchestration for Polymetrics
argument-hint: "<parent-issue-or-task>"
---

# Polymetrics Parent Orchestration

Task or parent issue:

$@

Run active parent issue orchestration for Polymetrics.

Required reading before action:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
- `.agents/agentic-delivery/contracts/pm-review-system.json`
- `.agents/agentic-delivery/contracts/pm-review-packet-template.md`
- `.agents/agentic-delivery/workflows/shepherd-validator.md`
- `.agents/agentic-delivery/contracts/pm-worker-handoff-template.md`
- `.agents/agentic-delivery/contracts/pm-code-review-disposition-template.md`
- `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md`

Operate as the live parent orchestrator in the main Pi session. Build the ready queue, create or
confirm the parent branch and parent PR, and delegate independent ready work through the
`subagent` tool using project agents from `.pi/agents/`:

- Dispatch `pm-gsd-worker` (mutating, `model: openai-codex/gpt-5.6-sol:high`, scoped to
  `read,bash,edit,write,grep,find,ls`) for each independent ready sub-issue with disjoint write
  scope. Give every mutating worker its own `cwd` (prefer a git worktree).
- Dispatch `pm-scout` (read-only, `model: openai-codex/gpt-5.6-sol:xhigh`) for reconnaissance
  sidecars.
- Dispatch `pm-verifier` (read-only, `model: openai-codex/gpt-5.6-sol:xhigh`) for exact-head
  verification.
- After verification, run `scripts/pm-review-system.py compile` for exact-base/head/tree closure,
  authority, typed bidirectional practical impact, semantic, scope, and packet coverage. Stop on
  unresolved impact, any graph/packet bound, or unsplittable context. Dispatch a fresh-context
  `pm-reviewer` (candidate read-only,
  `model: openai-codex/gpt-5.6-sol:xhigh`) per bounded packet, require complete responses, and run
  `synthesize` for exactly one PM-owned local-Codex result. Reviewers build impact first and may
  test hypotheses only through bounded disposable `scripts/pm-review-lab.py` copies; lab ambiguity,
  denial, inconclusive evidence, cleanup failure, or candidate drift blocks. Disposition every
  finding; any changed head requires verification, recompilation, fresh packet/lab evidence, and
  synthesis.
- Run independent Shepherd validation through `shepherd-validator.md` after review is clean and
  before integration.
- Run coupled/critical-path slices that cannot be isolated as `local_critical_path` in the main
  session via `/pm-gsd-loop`, never label an inline pass as `spawned`.

Pi runtime constraints:

- Use `agentScope: "both"` (or `"project"`) so project agents from `.pi/agents/` are visible.
  The default `"user"` scope does not load project agents.
- In non-interactive runs (`pi -p`), project agents are blocked unless you set
  `confirmProjectAgents: false`. Only set it to `false` after reviewing and trusting the project
  agents; otherwise run interactively.
- Give every mutating worker its own `cwd` (prefer a git worktree) so workers do not share the
  coordinator checkout. Read-only explorer/reviewer agents may share the coordinator checkout.
- Parallel mode is capped at 8 total tasks and 4 concurrent subprocesses per `subagent` call.
- Chain mode is capped at 8 sequential steps.
- Recursive subagent calls are blocked; the orchestrator is the only spawner.
- Persist `max_correction_rounds: 4` unless the parent contract supplies another bound, plus
  `rounds_by_range` keyed by exact base/candidate lineage. Exceeding the cap is
  `not_spawned_review_blocked` and a human gate; never reset it through a replacement PR.
- Read-only agents (`pm-scout`, `pm-reviewer`) request `grep`/`find`/`ls`. The parent Pi session
  must enable those tools with `--tools read,bash,edit,write,grep,find,ls,subagent`.

Use compact caveman-style status for progress and handoffs, but keep commands, tests, code,
security warnings, destructive-action warnings, and human gates exact.

Run `scripts/gsd doctor`, `scripts/gsd list`, and source discovery. If `programming-loop` is absent,
do not invoke or invent it. This `/pm-orchestrate` owner executes PLAN → RED → GREEN → REFACTOR →
VERIFY → REVIEW → INTEGRATE with durable state. Local review follows
`local-codex-review-loop.md` plus the deterministic packet compiler/synthesis contract; independent
trajectory validation remains separate and follows `shepherd-validator.md`.
Claude and GitHub Copilot are not required, requested, or fallback PM coverage.

Hard stops:

- Do not request, print, store, summarize, or invent secrets.
- Do not push to `main`.
- Do not merge a parent PR to `main` without human approval.
- Do not integrate until every local Codex finding has a written disposition and Shepherd returns
  `PROCEED` for the exact reviewed head.
- If no worker is spawned while ready work remains, record the exact `not_spawned_*` blocker and
  the next unblock action.
