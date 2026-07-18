# Phase 424 Prompts

## Kickoff snapshot

Task: Execute polymetrics-ai/cli#424 as the fourth serialized Phase 9 namespace worker for umbrella #407 and parent #397.

Command path:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research >/tmp/gsd-plan-phase-424-runtime-native-cobra.prompt
scripts/gsd prompt programming-loop init --phase 424-runtime-native-cobra --dry-run >/tmp/gsd-programming-loop-424-runtime-native-cobra.prompt
```

Result:

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research`: prompt generated (10739 bytes).
- `scripts/gsd prompt programming-loop init --phase 424-runtime-native-cobra --dry-run`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`). Manual fallback uses `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

Downstream artifact: `.planning/phases/424-runtime-native-cobra/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`.

Verification result: pass — red runtime/router tests captured; focused runtime/golden, full internal CLI, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, runtime help/docs/website parity, and diff guards passed.

Execution decision: `local_critical_path` — worker cwd/branch isolated; no subagent tool available; no delegation.
