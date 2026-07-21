# Pi Autonomous Orchestration Loop

Fully autonomous, resumable delivery loop. Given one prompt describing any problem (connector or
implementation), Shepherd researches, plans, creates issues, implements, verifies, reviews,
corrects, and integrates through bounded in-process Pi `AgentSession` roles. It can be stopped and
resumed after process or host interruption without losing authoritative progress.

This is the runtime-generic contract. The authoritative Pi implementation is the `/pm-shepherd`
extension under `.pi/extensions/shepherd/`; `.pi/prompts/pm-auto-loop.md` and the shell drivers are
transitional rollback surfaces until issue #471's canary and cutover pass. It composes the existing
`parent-issue-orchestration-loop.md`, `pi-active-orchestration-loop.md`, `gsd-universal-runtime-loop.md`,
`claude-review-loop.md`, and `code-review-disposition-template.md`.

## Roles → agents → models

| Stage role | Agent | Model | Provider |
|---|---|---|---|
| Orchestrator policy/controller | host code plus validation session | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |
| Web / API research | `pm-web-researcher` | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |
| Parent + task planning | `pm-planner` | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |
| Issue creation | typed host adapter plus `pm-issue-creator` proposal | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |
| Execute / correct | `pm-gsd-worker` | `openai-codex/gpt-5.6-sol` (`high`) | Codex |
| Verify | `pm-verifier` | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |
| Review + disposition | `pm-reviewer`, `pm-claude-review-disposition` | `openai-codex/gpt-5.6-sol` (`xhigh`) | Codex |

The controller is the only spawner; child sessions cannot create sessions recursively. Every role
is created through Pi's public `createAgentSession` API inside the current Pi process. The host
controller—not model prose—owns transitions, typed external mutations, persistence, and authority.
It persists state after every transition so any transition boundary is a safe resume point.

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
  REVIEW        (Orchestrator / pm-reviewer, on the sub-PR) → adversarial findings   ── GATE: must be clean ──
  CORRECT       (Codex  / pm-gsd-worker) if findings → fix → push  ┐
                (Orchestrator / pm-reviewer) re-review                    ┘ repeat ≤ max_correction_rounds
  INTEGRATE     merge sub-PR → parent branch; mark sub-issue complete
─────────────────────────────────────────────────────────────────────────────────────────
PARENT_FINALIZE parent PR coverage + disposition → durable exact-head human decision request
HUMAN_DECISION  wait for an allowlisted `/shepherd decide <id> approve-merge` comment
MERGE           revalidate exact head and gates, then wait for and observe the human-owned merge
COMPLETE        confirm GitHub/default-branch truth and close the durable run
```

RESEARCH is skipped entirely for a fully-specified `implementation` task (no external unknown) — the
`implementation` path stays first-class and light. Gates are hard: RESEARCH (connector) must be
`complete` before `PARENT_PLAN`; `VERIFY` must pass before `REVIEW`; `REVIEW` must be clean (every
finding fixed or dispositioned per `code-review-disposition-template.md`) before `INTEGRATE`. Branch
and PR creation (`PARENT_SETUP`, `SUB_BRANCH`, `SUB_PR_OPEN`) are idempotent—reconcile GitHub and
Git before creating anything. A parent merge is never inferred: Shepherd posts one durable
head-bound decision request, waits, accepts one allowlisted response, revalidates, and records that
the exact head is ready for the human-owned merge. Shepherd exposes no parent-to-`main` merge
operation. Until GitHub and the default branch authoritatively show that a human completed the
merge, the loop remains in observer-only `MERGE` rather than claiming success.

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
  "prompt": "<original problem prompt>",
  "problem_type": "connector | implementation",
  "stage": "INTAKE | RESEARCH | PARENT_PLAN | ISSUE_CREATE | PARENT_SETUP | TASK_LOOP | PARENT_FINALIZE | done | blocked | human_gate",
  "research": { "needed": false, "slug": "", "doc_path": "", "endpoints_found": 0, "writes_found": 0, "unclassified_endpoints": 0, "all_source_urls_present": false, "complete": false },
  "parent_issue": 0,
  "parent_branch": "",
  "parent_pr": 0,
  "subissues": [{ "number": 0, "title": "", "stage": "not_started", "branch": "", "sub_pr": 0, "deps": [], "write_scope": "" }],
  "guards": { "iteration": 0, "max_iterations": 200, "correction_rounds": {}, "max_correction_rounds": 4 },
  "terminal": null
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
- **Correction cap**: `max_correction_rounds` (default 4) per sub-issue. On exceed, mark the
  sub-issue `blocked` with the outstanding findings and stop for human review — never loop forever.
- **Iteration cap**: `max_iterations` (default 200) orchestrator turns per run as a hard backstop.
- **Loop safety**: stalled or repeating workers are detected via the reconciler's commit/stall
  check and re-dispatched or escalated, not silently retried.
- **Isolation**: every mutating worker gets its own `cwd` (sibling checkout or worktree). A
  mutating worker without isolation is recorded `not_spawned_isolation_missing` and not spawned.

## Human gates and durable waiting

- Requirements/scope/authority decisions are requested on the parent issue; exact-head review and
  merge decisions are requested on the relevant PR.
- Every request has an idempotency marker, request ID, allowed options, generation, allowlisted
  actor, and exact head when applicable. It is answered only by
  `/shepherd decide <request-id> <option>` on the bound issue or PR.
- Parent merge, any secret access, new dependency, token-scope change, destructive external action,
  production deploy, quality-gate reduction, correction-cap exhaustion, or a finding marked
  `Needs human` requires this durable wait/decision flow.
- Silence, emoji, CI success, review prose, or an agent rating never grants authority.

## Termination

The run ends successfully only when all sub-issues are integrated and verified, the exact-head
human merge decision was consumed, and GitHub plus the default branch prove the human-owned parent
merge. Consuming approval never authorizes Shepherd itself to merge or push to `main`.
Human-wait, blocked, and budget states are resumable terminal conditions for the current process,
not successful completion. Success is never assumed from a missing error.

## Runtime and migration policy

The authoritative runtime is `/pm-shepherd`, using in-process Pi `AgentSession` children and typed
host adapters. It must not depend on tmux, a Go Shepherd daemon, `pi-sub-agent`, or a second `pi`
process. The Go Shepherd issues and PRs are abandoned and superseded by #471.

`scripts/pi-shepherd-loop.sh`, `scripts/pi-auto-loop.sh`, and `scripts/claude-auto-loop.sh` remain
temporary rollback paths while the in-process controller is under construction. They must be
clearly labeled legacy and are deprecated only after crash/restart tests and the #397/#438 canary
pass. Do not delete their branches, worktrees, or traces as part of the replacement.

Billing hard rule: never route any role through OpenRouter or another pay-per-token gateway. Codex
roles stay on `openai-codex/gpt-5.6-sol` through the configured ChatGPT subscription path.

## Validator layer (Shepherd supervisor meta-agent)

The controller dispatches an independent validation `AgentSession` that judges whether the
orchestration is running correctly, step by step—see
`.agents/agentic-delivery/workflows/shepherd-validator.md`. After every orchestrator turn it scores
the `(state, action)` transition 1–5 on six dimensions (Anthropic trajectory rubric, geometric mean),
writes `VALIDATION.jsonl`, and emits a verdict — `PROCEED` / `RETRY` / `REVERT` (restore the last
checkpoint and replay the stage from that fork point) / `HALT` (hard-gate breach → human). It is
independent of the implementation worker (it re-derives truth through read-only typed evidence), so
a hallucinating worker cannot self-certify. Deterministic host gates override every rating and
catch skipped gates, no-op loops, stale evidence, and parallel-worker write conflicts.
