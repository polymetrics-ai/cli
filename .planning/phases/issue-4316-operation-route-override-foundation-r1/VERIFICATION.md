# Verification — Issue 4316

## Checklist

- [x] Resolver carries operation declaration source trace, base, version, and route without caller-controlled URLs.
- [x] Missing routing returns an explicit missing-foundation diagnostic before any provider I/O.
- [x] Conflicting declaration bases return the same diagnostic before provider I/O.
- [x] Direct read, direct write, binary download, binary upload, ETL, and reverse ETL compose their provider request through the shared resolver.
- [x] Five real Help Scout v3 direct-read operations pass their behavioral URL assertions against a configured `/v2` fixture base.
- [x] Source trace, canonical mapping, generated surface, fixtures, and website projection are current.
- [ ] No edit to `internal/connectors/defs/github/rate_limits.json`.
- [x] CLI/manual/website applicability documented and generated output checked.
- [ ] Targeted package tests, individual verification gates, build, GSD verification, and code review evidence recorded.

## Executed evidence

- `go test -timeout 20m ./internal/connectors/engine` — pass.
- `go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1` — pass.
- `go run ./cmd/connectorgen validate` — 552 connectors, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` — 552 connectors, no drift.
- `npm --prefix website run gen:website-data` and `git diff --check` — pass.
