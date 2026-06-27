# RELEASE-NOTES — Flow Engine (Phase 0)

## [Phase 0] Flow engine skeleton — sync + query steps

### New commands

- `pm flow plan [--file <path>] [--force] [--json]` — validate manifest, build DAG, execute
  read-only steps (sync + query), write ledger entries, return structured result.
- `pm flow preview [--file <path>] [--json]` — dry-run; returns execution plan without
  executing any steps.
- `pm flow run [--file <path>] [--force] [--json]` — full execution (equivalent to `plan`
  in Phase 0; action steps added in Phase 1).
- `pm flow status <name> [--json]` — show last run status for a named flow.
- `pm flow list [--json]` — list flow manifests under `.polymetrics/flows/`.

### New package

`internal/flow` — manifest parse/validate, DAG build, topological executor, checkpoint
store, sentinel errors.

### Behaviour

- Flow manifests are JSON (version 1) stored under `.polymetrics/flows/`.
- Steps are ordered by DAG derived from `in`/`out` table declarations.
- Cyclic dependencies are detected and reported with step IDs.
- Each step and the overall flow run are recorded in the ledger.
- Completed steps are checkpointed; re-runs skip already-successful steps.
- `--force` clears checkpoints and restarts from scratch.
- Concurrent runs on the same flow are prevented via a PID-checked file lock.
- Stale locks (crashed process) are automatically removed on next run.

### Zero new dependencies

All new code uses Go stdlib only. No entries added to `go.mod`.

### Breaking changes

None — new `pm flow` command is additive.

### Upgrade notes

Run `pm init` in existing projects to create `.polymetrics/flows/` and
`.polymetrics/locks/` directories (init is idempotent).
