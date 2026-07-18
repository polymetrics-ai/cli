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

Verification result: pass — PR #460 review-fix command refresh used `scripts/gsd doctor && scripts/gsd list`, `scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research >/tmp/gsd-plan-phase-424-runtime-native-cobra-review-fix.prompt`, and the same blocked `programming-loop` prompt path (`scripts/gsd: unknown GSD command: programming-loop`). Fixed Cobra/pflag parse usage mapping, runtime optional-service docs, and DragonflyDB/Temporal endpoint sanitization. Focused tests, full gates, docs/website/golden parity, and diff guards passed. User disallowed Claude/Copilot.

Execution decision: `local_critical_path` — worker cwd/branch isolated; no subagent tool available; no delegation.

## Positional-help correction snapshot

Task: Apply the bounded independent-review correction for issue #424 / PR #460 without reimplementing accepted work.

Identity: session `7050f706-72d2-47df-ac13-0b08979cc1ae`; model `openai-codex/gpt-5.6-sol`; thinking `high`; starting HEAD `8d696cd4c27fad6840e905917e7658e785fa5436`.

Command path: GSD doctor/list passed; plan-phase prompt generated at `/tmp/gsd-plan-phase-424-runtime-native-cobra-positional-help-correction.prompt`; `programming-loop` remained absent from the adapter registry, so the existing manual universal-loop fallback was used.

Downstream artifact: phase PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE plus focused changes in `internal/cli/cobra_router.go` and `internal/cli/runtime_cli_test.go`.

Verification result: pass — focused RED captured before production edits; focused runtime/router/golden/runtimecheck, built-binary positional text/JSON help, invalid-action usage, full gofmt/vet/tests/build/`make verify`, and diff/dependency/docs guards passed. No runtime services, credentials, dependency additions, new PR, or external review request.

Execution decision: `local_critical_path` — one known correction in the already isolated worker worktree; no independent subtask needed.
