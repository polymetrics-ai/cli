# Prompt trace — connector engine/direct-read policy migration

## Kickoff

Task: implement Issue-B focused connector-boundary migration after PR #605 by removing provider-specific GitHub date-range and repository-contents policies from shared engine/direct-read runtime.

## Required command path

- `scripts/gsd doctor` — healthy.
- `scripts/gsd list` — 69 commands available.
- `scripts/gsd prompt programming-loop --help` — not registered in this checkout; manual GSD universal-loop fallback used and recorded in plan/RUN-STATE.

## Downstream artifact

- `PLAN.md`
- `TDD-LEDGER.md`
- `VERIFICATION.md`
- `SUMMARY.md`
- `RUN-STATE.json`

## Verification result

Full local verification passed through `make verify`.
