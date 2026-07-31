# Verification Checklist — PostgreSQL parity wave 04 r1

## Required local gates

- [x] `go test ./internal/connectors/native/postgres`
- [x] focused CLI tests covering postgres command help / write plan
- [x] focused connectorgen/static conformance for `postgres`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify`
- [x] `git diff --check`

## Safety checks

- [x] No live DB/provider calls performed.
- [x] No secret values printed or committed.
- [x] No raw generic SQL/query/write command added.
- [x] Reverse ETL remains plan -> preview -> explicit approval -> execute.
- [x] Destructive `truncate_table` has typed destructive confirmation and no cascade/restart-identity escape.
- [x] API surface ledger counts match landed audit totals.
- [x] Captain-policy addendum appended idempotently to #3118-#3125 with truthful counts.

## Results log

### Pre-termination baseline

- `go test -count=1 ./internal/connectors/native/postgres` — pass.
- `go test -count=1 ./internal/cli -run TestPostgresCommandSurfaceHelpAndWritePlan` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass, 549 connectors, 0 findings; postgres 0 findings/warnings.
- `go test -count=1 ./internal/connectors/conformance -run TestConformance/postgres` — pass.
- `go build ./cmd/pm` — pass.
- `make connector-boundary` — pass, clean report.
- `go vet ./...` — pass.
- `go test -timeout 60m ./...` — pass. The first `make verify` attempts exposed the known hard `go test -timeout 20m` package-duration issue after full uncached CLI/certify tests; after updating golden transcripts and warming/passing those packages, final `make verify` completed successfully.
- `make verify` — pass (final run).
- `git diff --check` — pass.
- `gh-axi issue edit` for #3118-#3125 — pass; each body has one `captain-policy-addendum-postgres-destructive-confirmation-r1` marker with counts: total=263, implemented/covered=5, fixture-tested writes=5, blocked=205, excluded=53, certified=0.

### Resume rerun after typed MERGE correction

- `go test -count=1 ./internal/connectors/native/postgres -run 'TestPostgresWrite(BuildsParameterizedSQL|ValidationRejectsArbitrarySQLAndUnsafeIdentifiers|FixtureWriteDryRunAndExecute)'` — pass.
- `go test -count=1 ./internal/connectors/native/postgres` — pass.
- `go test -count=1 ./internal/cli -run TestPostgresCommandSurfaceHelpAndWritePlan` — pass.
- `go test -count=1 ./internal/connectors/conformance -run 'TestConformance/postgres'` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass, 549 connectors, 0 findings, 0 warnings.
- Closed-schema/raw escape audit script — pass, 0 violations; 5 commands, 5 write actions, 5 fixtures.
- `make verify` — pass.
- `git diff --check` — pass.
- `go build ./cmd/pm` — pass.
- `make connector-boundary` — pass, clean report.
