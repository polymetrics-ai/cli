# Prompts — Phase 415 OpenTelemetry metrics

## Kickoff snapshot

Task: Implement issue #415 `feat(obs): add OpenTelemetry metrics` as stacked PR to parent `feat/cli-architecture-v2` for parent issue #397.

GSD commands attempted:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 415-otel-metrics --skip-research >/tmp/gsd-plan-415.txt
scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415.txt
```

Downstream artifact: `.planning/phases/415-otel-metrics/PLAN.md`.
Verification result: planning complete; programming-loop command unavailable, manual fallback to `.pi/prompts/pm-gsd-loop.md` recorded.

## Review-fix snapshot

Task: Review-fix PR #461 / issue #415 on branch `feat/415-otel-metrics` head `8748a03ba60042bdc29bd9cce1acf7c3d0b286a3`; do not reset/discard/recreate; no Claude/Copilot.

GSD commands attempted:

```bash
scripts/gsd doctor
scripts/gsd list >/tmp/gsd-list-415-review-fix.txt
scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run >/tmp/gsd-loop-415-review-fix.txt
```

Downstream artifact: `.planning/phases/415-otel-metrics/PLAN.md` review-fix section.
Verification result: review-fix verified; focused tests, benchmark, full Go gates, `make verify`, `git diff --check`, and dependency diff review passed; programming-loop command unavailable, manual fallback remains active.

## Independent-review correction snapshot

Task: bounded correction for accepted findings in `/tmp/pm-397-review-415.log` on PR #461 / issue #415, starting HEAD `c6138292cfcc7205f7968a54b57a65f933a3c1fa`; session `153cfaabe3df4733a85717da46513786`; model `openai-codex/gpt-5.6-sol`; thinking `high`.

GSD commands attempted:

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt programming-loop init --phase 415-otel-metrics --dry-run
```

Downstream artifacts: correction sections in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`.
Verification result: correction verified and pushed; exact RED captured before production edits; focused race/live-export/endpoint/reconciliation/Temporal tests, 5-run benchmark, full Go gates, module checks, `make verify`, and reviewed-range whitespace checks passed. Implementation commit `ceb3a35ce13642a0d8c8ea0010272582202f8afd` is on the existing remote PR branch. Programming-loop command remains unavailable, so manual fallback remains recorded.

## Manual fallback prompt source

`.pi/prompts/pm-gsd-loop.md` loaded because `scripts/gsd` does not expose `programming-loop` in this checkout.
