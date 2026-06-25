# TEST PLAN: Agentic ETL Platform

## Unit Tests

- CLI unknown command, usage error, runtime error, and JSON error shape.
- Identifier, path, URL, and terminal sanitizer rejection cases.
- Connector manifest completeness and redaction.
- Skill generation output structure.
- ETL batch writer call behavior.

## Integration/Smoke Tests

- `make verify`
- `./poly docs generate --dir docs/cli`
- `./poly skills generate --dir <tmp>`
- Existing sample ETL and reverse ETL smoke flow.

## Deferred

- `POLYMETRICS_INTEGRATION=1 go test ./...` only when runtime services are available.
- Live GitHub tests only with explicit approval and non-secret token handling.
