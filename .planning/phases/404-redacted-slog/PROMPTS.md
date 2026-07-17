# PROMPTS — Issue #404

## Kickoff snapshot

Task: execute issue #404 under parent #397 on branch `feat/404-redacted-slog` from base `20475ddf`, scoped to redacted stdlib slog foundation, per-run JSONL routing/retention, vault.Get redaction registry, Temporal structured logger bridge, focused tests, and issue-local planning artifacts.

Required command path attempted:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 404 --skip-research
scripts/gsd prompt programming-loop init --phase 404 --dry-run
```

Downstream artifact: `.planning/phases/404-redacted-slog/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`.

Verification result: planning pending production red tests; `programming-loop` adapter command missing, manual GSD fallback recorded.

## Manual-GSD fallback prompt in effect

Follow `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` manually: plan before coding, capture red tests before production edits, implement minimal green slices, run focused and full gates, commit/push coherent green checkpoints, keep issue-local artifacts current, and stop for human gates.
