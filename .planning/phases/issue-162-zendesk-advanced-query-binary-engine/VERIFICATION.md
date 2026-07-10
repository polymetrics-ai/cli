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

Pending.

## Safety verification

- Binary policy must emit metadata only: no body, no base64, no destination writes.
- No credentialed Zendesk requests are run.
