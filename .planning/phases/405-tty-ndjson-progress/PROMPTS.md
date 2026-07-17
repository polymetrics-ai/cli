# PROMPTS — Issue 405 TTY gate and NDJSON progress

## Kickoff snapshot

Task: Execute polymetrics-ai/cli#405 as bounded mutating worker for parent #397.

Branch: `feat/405-tty-ndjson-progress`
Parent PR: #438
Base: `feat/cli-architecture-v2` at `d8e532eb20d24d982c09772ecae48abb3bb64271`
Write scope: CLI run options/global UI flags, TTY detection, event-to-stderr NDJSON wiring, `internal/ui/styles/**` foundation, focused docs/tests, and this phase directory.

## GSD prompt commands

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 405 --skip-research
scripts/gsd prompt programming-loop init --phase 405 --dry-run
```

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, focused red tests, implementation, docs parity, PR.

Verification result: pending. `programming-loop` prompt command failed with `scripts/gsd: unknown GSD command: programming-loop`; loaded `.pi/prompts/pm-gsd-loop.md` and running inline/manual universal loop.

## Acceptance snapshot

- TUI gate: stdout TTY and not `--json`, `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, or `TERM=dumb`.
- `RunWithOptions` exists; `Run` stays plain by construction.
- `--progress ndjson` emits sanitized progress events to stderr; stdout remains the final command output/envelope.
- Palette/glyph foundation degrades for no-color and ASCII constraints.
- CLI help/docs/website parity required.
