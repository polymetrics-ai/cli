# Verification: Intercom CLI Surface Metadata

## Focused Gates

```bash
gofmt -w cmd/connectorgen/intercom_api_surface_test.go
go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1
jq empty internal/connectors/defs/intercom/api_surface.json internal/connectors/defs/intercom/cli_surface.json .planning/phases/issue-165-intercom-cli-surface-metadata/RUN-STATE.json
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/engine -run 'Intercom|CLISurface' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
```

## Results

- Gofmt passed for the new Intercom API surface metrics test.
- Intercom API surface metrics test passed.
- JSON parse checks passed.
- Full connector definition validation passed: 547 connector(s) checked, 0 findings.
- Focused engine CLI surface load tests passed.
- `cmd/connectorgen` and `internal/connectors/engine` package tests passed.
- Intercom conformance passed.
- Focused connectorgen CLI surface tests passed.

## Broader Gates Before Handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test ./... -timeout=20m
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Status:

- `gofmt -w cmd internal` passed.
- `go vet ./...` passed.
- `go test ./...` with the default 10-minute package timeout timed out in `internal/connectors/certify` while running `TestWriteCreateFailureRecordsNoLeak`; this is a duration timeout, not a test assertion failure.
- Focused retry `go test ./internal/connectors/certify -run TestWriteCreateFailureRecordsNoLeak -count=1 -timeout=20m` passed in 41.571s.
- `go test ./... -timeout=20m` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed, including gofmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validation, smoke, lint, and connectorgen validation.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed through `make verify` and earlier focused run.

## CLI Help / Docs / Website Parity

- Runtime help: deferred to #166; #165 adds metadata only.
- Bare namespace behavior: deferred to #166.
- `pm <command> --help`: deferred to #166.
- `docs/cli/**`: deferred to #166 unless generated metadata checks require an update.
- `website/**`: deferred to #166 unless generated metadata checks require an update.
- Generated help/manual artifacts: deferred to #166.
- Connector `docs.md` updated in this slice to document the full API surface metadata without overclaiming executable writes/direct reads.
