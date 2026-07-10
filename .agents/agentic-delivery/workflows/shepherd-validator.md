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

The loop's durable state IS the reversible trace — no new instrumentation needed:

- `RUN.json` + `ORCHESTRATION-STATE.json` snapshots (one per turn, in `.planning/auto-loop/checkpoints/<turn>/`)
- `git log` / diffs on the parent + sub branches (each green slice is a commit = a revertible node)
- GitHub state (`gh issue/pr` — issues, PRs, review threads, dispositions)
- stage artifacts (`RESEARCH.{md,json}`, `PLAN.md`, `VERIFICATION.md`, review findings)
- `driver.log` (the ordered event log of orchestrator turns)

## What the validator does each turn (after the orchestrator advances one stage)

1. **Reconstruct the last transition** `(state_before → action → state_after)` from the trace.
2. **Score it 1–5 on each dimension** (independent of the orchestrator's own claim — verify against
   git/gh/artifacts, ground truth wins):
   - `correct_stage` — was the stage the orchestrator advanced the *right* one given true state?
   - `artifact_valid` — did the action produce the expected, complete artifact? (research doc
     `complete`/0-unclassified; plan maps every endpoint; verify actually ran the gates; review left
     a disposition on every finding)
   - `gates_respected` — no INTEGRATE before VERIFY+REVIEW green; no merge to `main`; no secret
     printed/stored; no new dep/destructive action without a human gate.
   - `real_progress` — the trace moved forward (a new commit / issue / PR / artifact), not a no-op or
     a repeat of the previous turn.
   - `no_hallucination` — the orchestrator's claimed state matches ground truth (ledger vs git/gh).
   - `no_conflict` — no two mutating workers touched the same file/branch/shared bundle file in
     parallel (the Shepherd supervisor's core job).
3. **`step_score`** = geometric mean of the six dimension scores. Append a line to
   `.planning/auto-loop/VALIDATION.jsonl` (schema below). **`trajectory_geomean`** = geometric mean
   of all `step_score`s so far.
4. **Emit a verdict** and write it to `.planning/auto-loop/VALIDATOR-VERDICT.json` for the driver:

   | Verdict | When | Driver action |
   |---|---|---|
   | `PROCEED` | `step_score ≥ 4` and no hard-gate breach | tag a checkpoint; continue |
   | `RETRY` | `2 ≤ step_score < 4`, recoverable (bad artifact, missed disposition, wrong dispatch) | re-run the orchestrator turn with the validator's `correction` appended |
   | `REVERT` | `step_score < 2` or repeated `RETRY` on the same stage | restore the last `PROCEED` checkpoint (state snapshot) and replay the stage from that fork point |
   | `HALT` | hard-gate breach (merge to main, secret exposure, destructive action, unresolved write-scope conflict) or `max_reverts` hit | stop; `terminal: blocked`; report for human |

5. Keep the verdict **specific and actionable** — the `reason` cites the exact trace evidence
   (commit SHA, issue/PR number, artifact path, the failing check), and `correction` tells the
   orchestrator precisely what to redo.

## Checkpoints, revert, and replay (Shepherd-style)

- After every `PROCEED`, the driver records a checkpoint: the current parent/sub branch SHAs + a copy
  of `RUN.json`/`ORCHESTRATION-STATE.json` under `.planning/auto-loop/checkpoints/<turn>/`.
- `REVERT` restores the last checkpoint's **local** state snapshot and re-dispatches the stage (a
  counterfactual replay from the fork point). Because every stage is idempotent and every slice is a
  committed node, replay never double-applies and never loses a good earlier slice.
- The validator never rewrites **pushed** git history or merges/force-pushes — revert is a local
  state restore + stage re-run. Anything requiring history surgery is a `HALT` for human.

## VALIDATION.jsonl entry (one per orchestrator turn)

```json
{
  "turn": 12,
  "stage_from": "EXECUTE",
  "action": "opened sub-PR #341 for surface-metadata slice",
  "stage_to": "SUB_PR_OPEN",
  "checks": { "correct_stage": 5, "artifact_valid": 4, "gates_respected": 5, "real_progress": 5, "no_hallucination": 5, "no_conflict": 5 },
  "step_score": 4.63,
  "trajectory_geomean": 4.41,
  "verdict": "PROCEED",
  "reason": "sub-PR #341 exists (gh), base=parent branch, Refs #<sub>+#<parent>; ledger matches.",
  "correction": null
}
```

## Guarantees

- **No step goes unjudged** — the validator runs after every orchestrator turn; the geometric mean
  means a single bad step drags the trajectory score down and triggers RETRY/REVERT rather than being
  masked by good steps.
- **Independent of the orchestrator** — the validator re-derives truth from git/gh/artifacts, so a
  hallucinating orchestrator can't self-certify.
- **Bounded** — `max_reverts` per stage and per run; a hard-gate breach is an immediate HALT.
- The validator is read-mostly: it writes only `VALIDATION.jsonl`, `VALIDATOR-VERDICT.json`, and
  checkpoint copies — never production code, never GitHub, never `main`.
