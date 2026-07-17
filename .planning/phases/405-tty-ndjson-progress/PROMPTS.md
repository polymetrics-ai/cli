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

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, focused red tests, implementation, docs parity, and stacked PR.

Verification result: local gates passed. `programming-loop` prompt command failed with `scripts/gsd: unknown GSD command: programming-loop`; loaded `.pi/prompts/pm-gsd-loop.md` and ran inline/manual universal loop.

## Acceptance snapshot

- TUI gate: stdout TTY and not `--json`, `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, or `TERM=dumb`.
- `RunWithOptions` exists; `Run` stays plain by construction.
- `--progress ndjson` emits sanitized progress events to stderr; stdout remains the final command output/envelope.
- Palette/glyph foundation degrades for no-color and ASCII constraints.
- CLI help/docs/website parity required.

## Final verification snapshot — 2026-07-17

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Verification result: pass; `go test ./...` included `internal/cli 173.250s`, `internal/connectors/certify 346.045s`; `make verify` included `internal/cli 174.571s`, `internal/connectors/certify 348.926s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`.

CLI parity result: pass for runtime help, bare `pm etl`/`pm flow`, invalid `pm flow bogus` exit 2 on initialized root, `docs/cli/**`, and website grep.
