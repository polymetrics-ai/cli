# Summary — Phase 408 flow/ETL dashboards

Status: in progress.

## Current state

- Required docs and skills loaded.
- GSD adapter healthy for `doctor`, `list`, and `plan-phase` prompt generation.
- `programming-loop` prompt is unavailable in `scripts/gsd`; manual universal-loop fallback recorded.
- Worker branch fast-forwarded from `5b603788` to parent `b77d8f49` before production edits.
- Issue-local phase artifacts created.
- EXECUTE resumed at `361a6bec0af1ed9cf84d5bdfdd10f16458d9da4d`; all 19 existing dirty entries adopted intact.
- Focused GREEN and focused race gates are recorded; full repository verification remains in progress.

## Delivered so far

- Issue-local GSD artifacts.
- RED tests for dual-TTY detection, flow/ETL dashboard activation and bypasses, dashboard frames, cancellation, layout/accessibility, sanitization/redaction, and bridge throttling.
- Minimal GREEN implementation:
  - stdin+stdout TTY detection (`RunOptions.StdinIsTerminal`, `DetectOptions.StdinTTY`);
  - `cmd/pm` auto mode while `cli.Run` stays plain;
  - `internal/ui/run` dashboard model, lifecycle-preserving throttle bridge, event-driven session, live inline refresh, and final scrollback frame;
  - `pm flow run` / `pm etl run` dual-TTY dashboards with parent/SIGINT cancellation propagated to engine contexts;
  - runtime help, docs/cli, and website parity updates.

## Next

1. Commit/push the focused green implementation checkpoint.
2. Run full repository tests, full race, and `make verify`; fix only issue-scoped regressions under RED → GREEN → REFACTOR.
3. Finalize truthful phase state and push verification checkpoint. Do not open a PR in this EXECUTE stage.

## Blockers / human gates

- No current human gate.
- New dependencies beyond ADR-0003 approved phase budget remain a hard stop.
- NTCharts remains unapproved and forbidden.
- Bubble Tea/teatest are absent from the live module and this EXECUTE instruction forbids new dependencies; literal teatest coverage remains an explicit verification gap, with deterministic headless model/session coverage used instead.
