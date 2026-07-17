# SUMMARY — Issue 405 TTY gate and NDJSON progress

Status: planning complete; production implementation not started.

## Planned delivery

- Deterministic `internal/ui` TTY gate per ADR 0003.
- `cli.RunWithOptions`; existing `Run` delegates with plain mode.
- Global `--plain`, `--no-input`, and `--progress ndjson` parsing.
- `events.NDJSON` progress wired to stderr only.
- `internal/ui/styles` palette/glyph foundation with no-color and ASCII fallback.
- Runtime help, docs/cli, and website parity.

## Current verification

- GSD adapter doctor/list passed.
- `plan-phase 405 --skip-research` prompt generated.
- `programming-loop` command unavailable; manual inline GSD fallback recorded.
- `verificationPassed=false` until implementation and full gates pass.

## Pending

- Red tests.
- Minimal implementation.
- Local gates.
- Stacked PR to `feat/cli-architecture-v2` using `Refs #405` and `Refs #397`.
