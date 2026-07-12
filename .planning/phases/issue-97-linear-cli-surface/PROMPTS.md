# Prompts — Issue #97 Linear CLI surface metadata

## Programming loop prompt attempt

Command used:

```bash
scripts/gsd prompt programming-loop init --phase issue-97-linear-cli-surface --dry-run
```

Downstream artifact: manual-GSD fallback recorded in this phase.

Verification result: unavailable (`scripts/gsd: unknown GSD command: programming-loop`).

## Local worker assignment

Worker: local orchestrator in `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-80-linear-cli-parity`.

Allowed write scope:

- `internal/connectors/defs/linear/cli_surface.json`
- `internal/connectors/engine/bundle_test.go`
- `.planning/phases/issue-97-linear-cli-surface/**`

Safety gates: no secrets, no credentialed checks, no new dependencies, no raw GraphQL or generic write tools, no reverse ETL execution.
