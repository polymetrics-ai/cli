# GSD programming-loop adapter fallback — issue 2997

Command attempted before production edits:

```bash
scripts/gsd prompt programming-loop init --phase issue-2997-gong-official-api-parity --dry-run
```

Observed result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

`scripts/gsd doctor` and `scripts/gsd list` passed, but this repo-local adapter registry does not expose the `programming-loop` prompt name in this worktree. Manual-GSD fallback is recorded in `.planning/phases/issue-2997-gong-official-api-parity/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`.
