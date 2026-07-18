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

## Manual fallback prompt source

`.pi/prompts/pm-gsd-loop.md` loaded because `scripts/gsd` does not expose `programming-loop` in this checkout.
