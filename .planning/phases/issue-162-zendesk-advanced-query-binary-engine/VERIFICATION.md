# Verification: Zendesk Advanced Query / Binary Engine

## Planned checks

```bash
go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadEmbeddedZendeskCLISurface' -count=1
go test ./internal/connectors/commandrunner -run 'DirectRead.*Binary|Zendesk' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
./pm zendesk binary show-attachment --help
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Results

Passed on 2026-07-10:

```bash
go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadZendeskDirectReadCommandCoverage|BundleLoadEmbeddedZendeskCLISurface' -count=1
go test ./internal/connectors/commandrunner -run 'BinaryManifest' -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
./pm zendesk binary show-attachment --help
./pm help zendesk
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Notes:

- `make verify` included `go mod tidy`, `go vet`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validation, smoke, golangci-lint, and `connectorgen validate`.
- `connectorgen validate`: 548 connectors checked, 0 findings.

## Safety verification

- Binary policy must emit metadata only: no body, no base64, no destination writes.
- No credentialed Zendesk requests are run.
