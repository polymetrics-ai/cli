# Prompts — Issue #80 Linear CLI parity parent

## GSD prompt snapshots

### Parent plan prompt

Command used:

```bash
scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research
```

Downstream artifact: `.planning/phases/issue-80-linear-cli-parity/PLAN.md`

Verification result: prompt generated successfully.

### Programming loop prompt attempt

Command used:

```bash
scripts/gsd prompt programming-loop init --phase issue-80-linear-cli-parity --dry-run
```

Downstream artifact: manual-GSD fallback recorded in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`.

Verification result: unavailable (`scripts/gsd: unknown GSD command: programming-loop`).

## Worker prompt status

No mutating worker prompt was dispatched because this harness lacks a Pi subagent tool. Spawn decision: `not_spawned_runtime_capability_missing`. Local critical-path implementation begins with issue #97.
