# Pi Autonomous Orchestration Loop

Fully automated, resumable, role-routed delivery loop. Given one prompt describing any problem
(connector or implementation), the loop plans, creates issues, implements, verifies, reviews,
corrects, and integrates. In the Codex-only profile, implementation runs at Sol/high and all
coordination/research/planning/verification/review roles run at Sol/xhigh. The loop can be stopped
and resumed at any point (including token exhaustion) without losing work.

This is the runtime-generic contract. The Pi adapter is `pm-auto-loop` (`.pi/prompts/pm-auto-loop.md`)
driven by `scripts/pi-auto-loop.sh`. It composes the existing
`parent-issue-orchestration-loop.md`, `pi-active-orchestration-loop.md`, `gsd-universal-runtime-loop.md`,
`local-codex-review-loop.md`, `shepherd-validator.md`, and `pm-code-review-disposition-template.md`.

## Roles → agents → models

| Stage role | Agent | Model | Provider |
|---|---|---|---|
| Orchestrator (main session) | — | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Web / API research | `pm-web-researcher` | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Parent + task planning | `pm-planner` | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Issue creation | `pm-issue-creator` | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Execute / correct | `pm-gsd-worker` | `openai-codex/gpt-5.6-sol` (high) | Codex |
| Verify | `pm-verifier` | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Fresh-context exact-head review | `pm-reviewer` | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |
| Independent trajectory validation | Shepherd validator context | `openai-codex/gpt-5.6-sol` (xhigh) | Codex |

The orchestrator is the only spawner (recursive `subagent` calls are blocked). The loop is driven
turn-by-turn by the orchestrator, which persists state after every transition so any turn is a
safe resume point.

## Stage machine

```
INTAKE          (Orchestrator) classify problem: connector | implementation; note whether external research is needed
RESEARCH        (Orchestrator / pm-web-researcher) → durable research doc  [ALWAYS for connector; for implementation only when INTAKE flags an external unknown]
                gate (connector): coverage self-check unclassified_endpoints==0, all_source_urls_present, complete==true
PARENT_PLAN     (Orchestrator / pm-planner mode=parent-plan) → parent + ordered sub-issues + dep graph  [consumes the research doc]
ISSUE_CREATE    (Codex  / pm-issue-creator) → gh issue create parent + subs (idempotent)
PARENT_SETUP    create parent branch feat/<N>-<slug> from main; open DRAFT parent PR → main (Refs #<N>; seed --allow-empty commit if no diff). Record parent_branch + parent_pr.
─ per ready sub-issue ──────────────────────────── TASK LOOP ─────────────────────────────
  TASK_PLAN     (Orchestrator / pm-planner mode=task-plan) → PLAN.md, TDD-LEDGER seed, VERIFICATION checklist
  SUB_BRANCH    create sub-branch off the parent branch, own cwd/worktree (mutating worker isolation)
  EXECUTE       (Codex  / pm-gsd-worker) → implement minimal green slices, commit per slice, push
  SUB_PR_OPEN   open sub-PR (base = parent branch; body: Refs #<sub> + Refs #<N>). Record sub_pr number.
  VERIFY        (Orchestrator / pm-verifier) → run gates → VERIFICATION.md   ── GATE: must pass ──
  REVIEW        (Orchestrator / fresh-context pm-reviewer) → exact-base/head findings   ── GATE: clean ──
  CORRECT       (Codex  / pm-gsd-worker) if findings → fix → verify → push  ┐
                (fresh-context pm-reviewer) exact-head re-review             ┘ repeat ≤ max_correction_rounds
                independent driver Shepherd validates the clean REVIEW transition: PROCEED required
  INTEGRATE     merge reviewed+validated exact head → parent branch; mark sub-issue complete
─────────────────────────────────────────────────────────────────────────────────────────
PARENT_FINALIZE (Orchestrator) parent PR coverage + disposition → human-ready gate (stop for human)
```

RESEARCH is skipped entirely for a fully-specified `implementation` task (no external unknown) — the
`implementation` path stays first-class and light. Gates are hard: RESEARCH (connector) must be
`complete` before `PARENT_PLAN`; `VERIFY` must pass before `REVIEW`; fresh-context local Codex
`REVIEW` must be clean (every finding fixed or dispositioned per
`pm-code-review-disposition-template.md`) and independent Shepherd must return `PROCEED` before
`INTEGRATE`. Branch
and PR creation (`PARENT_SETUP`, `SUB_BRANCH`, `SUB_PR_OPEN`) are idempotent — check `gh pr list`/`gh
issue list`/`git branch` and reuse what exists. Merges to `main` and final human-ready are human
gates — the loop stops and reports there.

## Durable state (source of truth for resume)

The loop never trusts session memory for progress. On every transition it writes durable state,
and on every wake it reconstructs progress from these four sources (ground truth wins over a stale
ledger):

1. **`ORCHESTRATION-STATE.json`** — the ledger, shaped by
   `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`, plus a per-sub-issue
   `stage` field from this loop:
   `not_started → task_planned → sub_branch → executing → sub_pr_open → verify_pending → review_pending → correcting → integrating → complete` (or `blocked`).
2. **GSD artifacts** per phase/sub-issue: presence + completeness of `PLAN.md`, `TDD-LEDGER.md`,
   `VERIFICATION.md`, `SUMMARY.md`.
3. **git** — branches and commits (how far the worker got; each green slice is a commit).
4. **GitHub** — issue existence/state, PR existence/state, review comments, and disposition replies.

Run-level state lives in `.planning/auto-loop/RUN.json`:

```json
{
  "schema_version": "canonical_v2",
  "prompt": "<original problem prompt>",
  "problem_type": "connector | implementation",
  "stage": "INTAKE | RESEARCH | PARENT_PLAN | ISSUE_CREATE | PARENT_SETUP | TASK_LOOP | PARENT_FINALIZE | done | blocked | human_gate",
  "research": { "needed": false, "slug": "", "doc_path": "", "endpoints_found": 0, "writes_found": 0, "unclassified_endpoints": 0, "all_source_urls_present": false, "complete": false },
  "parent_issue": 0,
  "parent_branch": "",
  "parent_pr": 0,
  "subissues": [{ "number": 0, "title": "", "stage": "not_started", "branch": "", "sub_pr": 0, "deps": [], "write_scope": "", "candidate_lineage": "" }],
  "correction_budget": { "max_correction_rounds": 4, "rounds_by_range": {} },
  "guards": { "iteration": 0, "max_iterations": 200 },
  "terminal": null,
  "human_gate_kind": null
}
```

## The reconciler (assess-the-stage — runs first on every wake)

Before doing any work, the orchestrator computes the true stage so a resumed run continues exactly
where it stopped (including after token exhaustion mid-task):

1. Load `RUN.json` and `ORCHESTRATION-STATE.json` → candidate stage per level.
2. Verify each candidate against ground truth and correct it:
   - `RESEARCH` claimed but the research doc is missing, `complete:false`, or (connector) any endpoint
     row lacks a `source_url` / `unclassified_endpoints > 0` → **resume RESEARCH** (idempotent; the
     researcher overwrites its own doc). Only `complete` research advances to `PARENT_PLAN`.
   - `PARENT_PLAN` claimed but no `PLAN.md`/decomposition → redo `PARENT_PLAN`.
   - `ISSUE_CREATE` claimed but a plan item has no GitHub issue → resume `ISSUE_CREATE` (idempotent; reuse existing).
   - `PARENT_SETUP` claimed but the parent branch or draft parent PR is missing on GitHub → resume
     `PARENT_SETUP` (idempotent: reuse an existing branch/PR; never open a duplicate).
   - sub-issue `sub_branch`/`executing`/`correcting` but the worker's last commit is behind its
     `PLAN.md`, or no handoff exists, or the worker is stalled (no expected-branch commit within
     `executor.stall_threshold_minutes`) → treat the worker as dead and **re-dispatch EXECUTE/CORRECT from the last commit**.
   - sub-issue `executing` with commits pushed but no recorded `sub_pr` and none found via `gh pr list`
     → resume `SUB_PR_OPEN`; if a matching sub-PR already exists, adopt its number.
   - `verify_pending` with a passing `VERIFICATION.md` → advance to `REVIEW`.
   - `review_pending` with all review threads dispositioned and clean → advance to `INTEGRATE`.
   - `integrating` with the sub-PR already merged → mark `complete`.
3. Write the reconciled state back, then continue the loop from the earliest non-complete stage.

Because every green slice is a committed checkpoint and every stage transition is persisted, the
worst case after any interruption is re-running one stage for one sub-issue — no work is lost and
nothing is double-applied (issue creation and merges are checked for idempotency first).

## Guards (baked in, not advisory)

- **Budget**: `scripts/pi-auto-loop.sh` runs under a per-run token/cost ceiling. When the ceiling
  is hit the orchestrator finishes the current durable transition, writes `terminal: "budget"`, and
  exits cleanly; re-running resumes from the reconciler.
- **Correction cap**: `correction_budget.max_correction_rounds` (default 4) applies to each stable
  exact-base/candidate lineage key in `correction_budget.rounds_by_range`. Create the candidate
  lineage identifier once when adopting the first candidate and preserve it across replacement PRs
  and changed heads; append head history instead of resetting the counter. On exceed, mark the
  sub-issue `blocked` with the outstanding findings, write `RUN.json.terminal = "human_gate"` plus
  sibling `RUN.json.human_gate_kind = "correction_cap_exceeded"`, and stop for a blocked human
  decision — never loop forever.
  A resumed pre-v2 record may read `guards.correction_rounds` once as read-only legacy input, map it
  into the stable lineage counter, and then write only the canonical `correction_budget` shape.
- **Iteration cap**: `max_iterations` (default 200) orchestrator turns per run as a hard backstop.
- **Loop safety**: stalled or repeating workers are detected via the reconciler's commit/stall
  check and re-dispatched or escalated, not silently retried.
- **Isolation**: every mutating worker gets its own `cwd` (sibling checkout or worktree). A
  mutating worker without isolation is recorded `not_spawned_isolation_missing` and not spawned.

## Hard stops (human gates — the loop stops and reports)

- Merge to `main`, and marking the parent PR human-ready.
- Any secret request/print/store, new dependency, token-scope change, destructive external action,
  production deploy, or quality-gate reduction.
- Correction cap exceeded, or a finding marked `Needs human`.

## Termination

The run ends only when: all sub-issues are `complete` and the parent reaches the human-ready gate
(`terminal: "human_gate"`, `human_gate_kind: "parent_ready"`); the correction cap is exceeded
(`terminal: "human_gate"`, `human_gate_kind: "correction_cap_exceeded"`, blocked human decision);
another hard stop/block is hit (`terminal: "blocked"`); or the budget ceiling is reached
(`terminal: "budget"`, resumable). Canonical producers always write `schema_version: "canonical_v2"`
and both human-gate fields. Success is not assumed from a missing error —
it is asserted from `ORCHESTRATION-STATE.json` + GitHub + git agreeing that every sub-issue is
integrated and verified.

## Canonical PM driver

Use the Codex-only PM route: `scripts/pi-shepherd-loop.sh` with `.pi/prompts/pm-auto-loop.md` (or
`/pm-connector-loop` for connector specialization). The orchestrator, isolated workers,
fresh-context local reviewer, and independent validator run on Codex through `pi` and the project
agent frontmatter. Requires the project `pi-sub-agent` extension.

If the GSD registry lacks `programming-loop`, `/pm-orchestrate` owns the stage machine above; never
invent the command or downgrade to an advisory fallback. Local code review follows
`local-codex-review-loop.md`; trajectory validation follows `shepherd-validator.md`.

The old `scripts/claude-auto-loop.sh` driver remains only for historical trace replay and is not a
current or forward PM orchestration or review route. Do not request GitHub-hosted AI reviewers.
Never route a PM role through OpenRouter or another pay-per-token gateway.

## Validator layer (Shepherd supervisor meta-agent)

Above the orchestrator sits an independent **Shepherd-style validator** that judges whether the
orchestration is running correctly, step by step — see
`.agents/agentic-delivery/workflows/shepherd-validator.md`. After every orchestrator turn it scores
the `(state, action)` transition 1–5 on six dimensions (Anthropic trajectory rubric, geometric mean),
writes `VALIDATION.jsonl`, and emits a verdict — `PROCEED` / `RETRY` / `REVERT` (restore the last
checkpoint and replay the stage from that fork point) / `HALT` (hard-gate breach → human). It is
independent of the orchestrator (re-derives truth from git/gh/artifacts), so a hallucinating
orchestrator cannot self-certify, and it is the layer that catches skipped gates, no-op loops, and
parallel-worker write conflicts before they compound.
