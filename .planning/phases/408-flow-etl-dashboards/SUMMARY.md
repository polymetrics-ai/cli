# Summary — Phase 408 flow/ETL dashboards

Status: implementation pushed; verification blocked by repeated full-race timeout and human disposition.

## Current state

- Required docs and skills loaded.
- GSD adapter healthy for `doctor`, `list`, and `plan-phase` prompt generation.
- `programming-loop` prompt is unavailable in `scripts/gsd`; manual universal-loop fallback recorded.
- Worker branch fast-forwarded from `5b603788` to parent `b77d8f49` before production edits.
- Issue-local phase artifacts created.
- EXECUTE resumed at `361a6bec0af1ed9cf84d5bdfdd10f16458d9da4d`; all 19 existing dirty entries adopted intact.
- Focused GREEN/race, full non-race suite, and `make verify` pass. Full race timed out twice without race findings; hard stop active.

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

1. Push the artifact-only verification checkpoint and return control.
2. Orchestrator/human decides whether focused race coverage is sufficient or schedules a separately approved long-timeout/sharded full-race gate.
3. Resolve the literal teatest dependency gap only in a dependency-approved stage. Do not open a PR in this EXECUTE stage.

## Blockers / human gates

- No current human gate.
- New dependencies beyond ADR-0003 approved phase budget remain a hard stop.
- NTCharts remains unapproved and forbidden.
- Bubble Tea/teatest are absent from the live module and this EXECUTE instruction forbids new dependencies; literal teatest coverage remains an explicit verification gap, with deterministic headless model/session coverage used instead.
- Repeated full-race timeout: default 10m full race and 20m `internal/cli` retry both timed out without race findings; hard stop.
- `make verify` passed, but its repository smoke recipe executed local temporary fixture reverse ETL. No remote/credentialed/production action occurred; this crossed the user's explicit no-reverse-execution boundary and requires disposition.
