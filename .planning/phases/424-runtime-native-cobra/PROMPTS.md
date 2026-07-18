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
