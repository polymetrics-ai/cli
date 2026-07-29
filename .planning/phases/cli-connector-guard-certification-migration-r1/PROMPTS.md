# Prompts — Connector Guard Issue C Certification Migration

## GSD commands

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt programming-loop init --phase cli-connector-guard-certification-migration-r1 --dry-run
scripts/gsd prompt plan-phase cli-connector-guard-certification-migration-r1 --skip-research
```

`programming-loop` is not registered in the current repo-local adapter, so this phase uses the documented manual-GSD fallback and records it in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`.
