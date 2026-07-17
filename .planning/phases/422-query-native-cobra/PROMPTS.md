# Phase 422 Prompts

## Kickoff snapshot

Task: Execute polymetrics-ai/cli#422 as the second serialized Phase 9 namespace worker for umbrella #407 and parent #397.

Command path:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 422 --skip-research >/tmp/gsd-plan-phase-422.prompt
scripts/gsd prompt programming-loop init --phase 422 --dry-run >/tmp/gsd-programming-loop-422.prompt
```

Result:

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt plan-phase 422 --skip-research`: prompt generated.
- `scripts/gsd prompt programming-loop init --phase 422 --dry-run`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`). Manual fallback uses `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

Downstream artifact: `.planning/phases/422-query-native-cobra/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`.

Verification result: pending red tests and implementation gates.

Execution decision: `local_critical_path` — worker cwd/branch isolated; no subagent tool available; no delegation.
