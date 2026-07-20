# Issue #397 orchestration prompt snapshots

## Pi 5.6 Sol routing correction — 2026-07-21

Source directive:

```text
Change the Pi parent orchestrator and all pm workers to openai-codex/gpt-5.6-sol. Use high for
implementation and xhigh for every other role. Continue the existing CLI Architecture v2 program
through a Shepherd-supervised, maximally parallel GSD/TDD loop until the final parent PR is ready
for human review; never merge the parent PR to main.
```

Execution contract:

- Parent issue #397, branch `feat/cli-architecture-v2`, draft parent PR #438.
- Active parent orchestrator contract plus GSD universal programming loop.
- Required skills loaded: `gsd-programming-loop`; `caveman` for compact long-running handoffs.
- `scripts/gsd doctor` passed. The adapter registry still lacks `programming-loop`, so the recorded
  manual universal-loop fallback was used for this bounded routing/driver correction.
- RED before routing/driver edits; GREEN role-routing and Shepherd self-tests; REFACTOR explicit
  project/session/driver policy; full local verification.
- Preserve the existing dirty #408 isolated worktree and resume it; never reset, clean, recreate,
  or dispatch a duplicate worker.
- Preserve Phase 437 pending intake as planning-only until separately authorized.
- Keep #419 at its explicit optional-dependency human gate.

Downstream artifacts:

- `.pi/settings.json`, `.pi/agents/*.md`, `.pi/prompts/*.md`
- `scripts/pi-auto-loop.sh`, `scripts/pi-shepherd-loop.sh`
- `scripts/tests/pi-model-routing.sh`, `scripts/tests/pi-shepherd-loop-verdict-guard.sh` (including
  a main-loop failed-validator regression)
- `.planning/config.json` and the issue #397 orchestration artifacts

Verification result: `make verify` passed, including model routing and both Shepherd guard tests.
