# Prompts — Phase 408 flow/ETL dashboards

## Kickoff snapshot

Command path:

```bash
scripts/gsd prompt plan-phase 408 --skip-research
scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run
```

Result:

- `plan-phase` prompt generated successfully at `/tmp/gsd-plan-408.txt`.
- `programming-loop` prompt unavailable: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual universal-loop fallback active per `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

Downstream artifact:

- `.planning/phases/408-flow-etl-dashboards/PLAN.md`
- `.planning/phases/408-flow-etl-dashboards/TDD-LEDGER.md`
- `.planning/phases/408-flow-etl-dashboards/VERIFICATION.md`
- `.planning/phases/408-flow-etl-dashboards/SUMMARY.md`
- `.planning/phases/408-flow-etl-dashboards/RUN-STATE.json`

Verification result:

- Original implementation reached focused/full non-race green, but literal Bubble Tea v2/teatest coverage was absent.
- Preserved: full race timed out at 10m; targeted `internal/cli` race timed out at 20m without race findings.
- Prior `make verify` passed but crossed the narrower worker dispatch boundary through a local temporary reverse-smoke fixture; sequence remained plan → preview → approval → execute with no credential, remote, production, or persistent write.
- Shepherd correction complete at implementation commit `c70ecf64`; `execute_complete=false`; independent VERIFY pending.

## Shepherd RETRY snapshot — 2026-07-20

Command path:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run
```

Result: doctor passed; adapter still reports `scripts/gsd: unknown GSD command: programming-loop`, so the already-recorded manual universal-loop fallback remains active.

Downstream artifact:

- Correction slice in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`.
- Real Bubble Tea v2 model/program and `teatest/v2` RED→GREEN evidence recorded in issue TDD/verification artifacts and implementation commit `c70ecf64`.
- Delegated shared-parent state will be synchronized to live parent head `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f` without claiming parent verification or review.

Verification result: strict RED captured with `go test ./internal/ui/run -run '^TestBubbleTeaV2ModelAndTeatestProgram$' -count=1`; setup failed because `charm.land/bubbletea/v2` was not required. After exact authorized pins, the direct interface RED failed because `*Model` lacked `Init`; GREEN then passed through real `teatest/v2`. Focused/full non-race/focused-race/module gates pass at `c70ecf64`. `execute_complete=false`; no CORRECT-stage `make verify`, full race, independent VERIFY, REVIEW, or INTEGRATE.
