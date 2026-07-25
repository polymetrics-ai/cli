# Shepherd Validator — supervisor meta-agent over the autonomous loop

A validator layer **above** the orchestrator. The orchestrator runs the delivery loop; the
Shepherd validator independently judges whether that orchestration is running *correctly*, step by
step, and can halt, retry, or revert-and-replay when it isn't.

Modeled on **Shepherd** (Northeastern/Stanford, *Programmable Meta-Agents via Reversible Agentic
Execution Traces*, arXiv:2605.10913): treat the execution as a first-class, reversible, Git-like
**trace**, and put a **supervisor meta-agent** over it that inspects and, when needed, forks/replays
from a divergence point. Scoring follows the 2026 trajectory standard (**Anthropic rubric**): score
each `(state, action)` pair 1–5, combine with a **geometric mean** so one bad step can't be averaged
away, and judge **every** step — not just the final result.

## Why this exists

The orchestrator already self-reconciles (it re-derives its stage from durable state each turn). The
validator is an **independent second agent** that judges the orchestrator from the outside — catching
the failure the orchestrator can't see in itself: hallucinated progress, a skipped gate, two workers
colliding on one file, a stuck no-op loop, or a silent quality drop. Shepherd's own headline is
exactly this: a supervisor over parallel coding agents lifted pass rate 28.8% → 54.7%.

## The trace (what the validator reads)

The loop's durable state IS the reversible trace:

- `RUN.json` + `ORCHESTRATION-STATE.json` snapshots (one per turn, in `.planning/auto-loop/checkpoints/<turn>/`)
- driver-owned `SHEPHERD-REQUEST.json`, which binds `exact_base_sha`, `exact_head_sha`,
  `exact_head_tree`, `candidate_lineage`, and `synthesis_sha256`
- the authenticated clean local-Codex synthesis artifact whose exact bytes produce
  `synthesis_sha256`; its base/head/tree must match the request and current clean worktree
- `git log` / diffs on the parent + sub branches (each green slice is a commit = a revertible node)
- GitHub state (`gh issue/pr` — issues, PRs, review threads, dispositions)
- stage artifacts (`RESEARCH.{md,json}`, `PLAN.md`, `VERIFICATION.md`, review findings)
- `driver.log` (the ordered event log of orchestrator turns)

A branch name, mutable PR ref, prose SHA, prior verdict, or matching-but-stale state file is not an
identity. Before invoking Shepherd, the driver must verify all five request fields against the
canonical review state, authenticated compiler record, synthesis JSON, current `HEAD^{tree}`, merge
base, and a clean worktree. Missing, malformed, dirty, stale, divergent, non-clean, or internally
disagreeing evidence blocks validation. A changed head/tree or synthesis byte invalidates the
request and every prior Shepherd result.

## What the validator does for an eligible review transition

Shepherd is downstream of exact-head local-Codex review. The driver invokes the independent
validator **only after** deterministic review compilation is authenticated and the exact-head
synthesis is `clean`. Preparatory orchestration may continue, but INTEGRATE, parent-readiness
mutation, and successful `done`/`human_gate` terminal handling remain unavailable until this gate
returns a bound `PROCEED`. The validator reconstructs and judges the full trajectory leading to that
review transition rather than silently skipping earlier actions.

1. **Load the immutable request first.** Read driver-owned `SHEPHERD-REQUEST.json`; reject a missing
   field or any mismatch in `exact_base_sha`, `exact_head_sha`, `exact_head_tree`,
   `candidate_lineage`, or `synthesis_sha256`. Recheck clean current identity before writing a
   verdict. Never repair, replace, or reinterpret the request.
2. **Reconstruct the transition** `(state_before → action → state_after)` and its preceding
   trajectory from the initial/last-good checkpoint, current ledgers, ordered driver log, and ground
   truth.
3. **Score it 1–5 on each dimension** (independent of the orchestrator's own claim — verify against
   git/gh/artifacts; ground truth wins):
   - `correct_stage` — was the stage the orchestrator advanced the *right* one given true state?
   - `artifact_valid` — are verification, authenticated clean synthesis, dispositions, and all
     expected artifacts complete and bound to the request?
   - `gates_respected` — no INTEGRATE before VERIFY + clean exact-head synthesis + Shepherd; no
     unapproved dependency, auth-scope change, destructive/admin action, production deploy,
     credentialed connector check, reverse-ETL action outside plan → preview → approval → execute,
     generic shell/HTTP/SQL write, quality-gate reduction, parent-readiness mutation, or merge to
     `main`; no secret printed or stored.
   - `real_progress` — the trace moved forward (a new commit / issue / PR / artifact), not a no-op or
     repeat.
   - `no_hallucination` — claimed state matches ground truth (ledger vs git/gh).
   - `no_conflict` — no two mutating workers touched the same file/branch/shared bundle file in
     parallel.
4. **`step_score`** = geometric mean of the six dimension scores. Append exactly one bound line to
   `.planning/auto-loop/VALIDATION.jsonl` (schema below). **`trajectory_geomean`** = geometric mean
   of all `step_score`s so far.
5. **Emit one bound verdict** in `.planning/auto-loop/VALIDATOR-VERDICT.json`:

   | Verdict | When | Driver action |
   |---|---|---|
   | `PROCEED` | `step_score ≥ 4`, exact identity still matches, and no hard-gate breach | tag a checkpoint; permit the next non-merge driver transition |
   | `RETRY` | `2 ≤ step_score < 4`, recoverable (bad artifact, missed disposition, wrong dispatch) | re-run with the validator's correction |
   | `REVERT` | `step_score < 2` or repeated `RETRY` on the same stage | restore a verified checkpoint and replay |
   | `HALT` | any identity drift, hard-gate breach, unsafe/missing rollback point, or exhausted revert budget | stop blocked for a human |

6. Keep the verdict **specific and actionable** — `reason` cites exact trace evidence and
   `correction` says precisely what to redo. The validator never merges, marks a parent PR ready,
   changes GitHub state, or edits the candidate.

## Checkpoints, revert, and replay (Shepherd-style)

- Before the first orchestrator/validator turn, the driver creates and verifies an `initial`
  checkpoint containing `HEAD.sha` plus either `RUN.json` or an explicit `RUN.absent` marker. A
  first-turn `REVERT` without that checkpoint is a fail-closed `HALT`, never a no-op continuation.
- After every bound `PROCEED`, the driver records a checkpoint under
  `.planning/auto-loop/checkpoints/<turn>/`: a copy of `RUN.json` plus `HEAD.sha` (the worktree HEAD —
  the **fork point**). `ORCHESTRATION-STATE` ledgers are rebuilt from ground truth on replay.
- `REVERT` does two things, because restoring bookkeeping alone would not undo the bad step (the next
  RECONCILE re-derives forward from ground truth):
  1. restores `RUN.json` to the verified last-good checkpoint (or removes it when the initial
     checkpoint records `RUN.absent`), and
  2. writes `.planning/auto-loop/REVERT-CLEANUP.json` = `{ good_fork_sha, diverged_head_sha, instruction }`.
     The next orchestrator turn reads it during RECONCILE and **actually undoes the diverged commits**:
     `git reset` for **local-only** (unpushed) commits after the fork point, or a **revert-forward**
     commit for already-pushed ones — per its own gates. Then it replays the stage from the fork point.
- The **validator itself never** rewrites git history, force-pushes, or merges. Reverting real
  side-effects is delegated to the orchestrator (which owns branch/PR state); anything that would need
  history surgery on a shared/pushed branch, or that the orchestrator can't safely undo, is a `HALT`
  for human. Because every stage is idempotent and every slice is a committed node, the replay after
  cleanup never double-applies and never loses a good earlier slice.

## VALIDATION.jsonl entry (one per eligible exact review transition)

```json
{
  "schema_version": "polymetrics.ai/shepherd-validation/v1",
  "turn": 12,
  "exact_base_sha": "<40-hex>",
  "exact_head_sha": "<40-hex>",
  "exact_head_tree": "<40-hex>",
  "candidate_lineage": "<stable exact-base/candidate lineage>",
  "synthesis_sha256": "<64-hex hash of the authenticated clean synthesis bytes>",
  "stage_from": "REVIEW",
  "action": "synthesized exact-head local-Codex review clean",
  "stage_to": "REVIEW_CLEAN",
  "checks": { "correct_stage": 5, "artifact_valid": 4, "gates_respected": 5, "real_progress": 5, "no_hallucination": 5, "no_conflict": 5 },
  "step_score": 4.63,
  "trajectory_geomean": 4.41,
  "verdict": "PROCEED",
  "reason": "clean synthesis and trajectory match the exact request; no configured human gate was crossed.",
  "correction": null
}
```

`VALIDATOR-VERDICT.json` uses schema `polymetrics.ai/shepherd-verdict/v1` and repeats the same five
identity fields verbatim. The driver rejects a verdict unless a new final `VALIDATION.jsonl` entry
also repeats them and agrees on the verdict. Prose citations are supplementary, never identity.

## Guarantees

- **No successful terminal or integration bypass** — a trajectory is judged only after authenticated
  clean synthesis, and all actions leading to that transition are reconstructed. Without a fresh
  bound `PROCEED`, terminal success and integration remain blocked.
- **Independent and exact** — the validator re-derives truth, while the driver rechecks identity
  before and after the validator. A stale, dirty, malformed, or self-certified result cannot pass.
- **Complete human gates** — every configured gate class above is a hard `HALT`; an unapproved
  parent-ready transition is not success.
- **Bounded liveness** — validator process and descendants run under a hard watchdog. Nonzero exit,
  timeout, watchdog error, missing append, or exhausted revert budget discards the verdict and cannot
  succeed.
- **Read-mostly and no merge** — the validator writes only `VALIDATION.jsonl` and
  `VALIDATOR-VERDICT.json`. The driver writes request/checkpoint/cleanup bookkeeping. Neither edits
  production code, mutates GitHub, rewrites history, force-pushes, marks ready, or merges any branch.
