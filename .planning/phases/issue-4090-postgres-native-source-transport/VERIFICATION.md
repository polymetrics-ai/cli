# Verification — Issue #4090

## Acceptance checklist

- [x] PostgreSQL's declared source preflights through the real transport registry.
- [x] `full_overwrite` and `full_append` emit bounded typed records and a valid checkpoint.
- [x] A missing descriptor refuses before I/O.
- [x] A wrong executor family refuses before I/O.
- [x] An unregistered declared executor refuses before I/O.
- [x] Focused unit and package race checks pass.
- [x] The PostgreSQL dbtest harness passes against an explicit Docker-or-Podman
      Unix endpoint and its output shows rows, emitted identity/schema, and checkpoint.
- [x] Required individual repository gates pass.
- [x] Manual inline `verify-work` and `code-review` records are complete.

## Safety audit

- [x] No secret value appears in test output, traces, docs, commit, or PR body.
- [x] No raw SQL/configured SQL path or direct source-to-destination route exists.
- [x] `internal/connectors/engine/bundle.go` is unchanged.
- [x] Changed paths remain PostgreSQL-only plus planning evidence.

## Executed local gates

All commands below passed during the final post-review verification run.

```text
go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/synctransport
go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres
go vet ./internal/connectors/native/postgres ./internal/synctransport
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
  go test -tags=databaseintegration -timeout 20m -count=1 \
  -run '^TestPostgresDynamicTypedCatalogUsesLiveMetadata$' -v \
  ./internal/connectors/native/postgres
```

The live run emitted five exact rows in three bounded pages for each full
mode. Its emitted source identity, typed catalog fingerprint, repeatable-read
snapshot barrier, and dedupe identity are retained verbatim in
`traces/live-source-green.txt`.
