# RUNBOOK / Rollback — GitHub Native Package + Data-Driven Registry

## Nature of change
Pure internal refactor: move GitHub connector into `internal/connectors/github/` and make the
registry self-registering. No schema, no data, no dependency, no external behavior change.

## Verify
- `make verify` green.
- `./pm connectors inspect github --json` → kind "Connector".
- `./pm connectors list` includes `github` (and `source-github` alias).

## Rollback
- Git revert the phase commits (no migrations to undo).
- If only the registry change is suspect: re-add `r.Register(Github{})` and remove the blank
  import in `registry_gen.go`; the package can coexist temporarily.

## Operational notes
- No runtime config flags introduced.
- No background jobs, queues, or external calls added.
