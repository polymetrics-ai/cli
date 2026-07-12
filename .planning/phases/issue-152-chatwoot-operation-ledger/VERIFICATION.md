# Verification: Chatwoot Operation Ledger

## Completed targeted gates

```bash
go test ./cmd/connectorgen -run ChatwootAPISurfaceOperationLedgerMetrics -count=1
go test ./cmd/connectorgen -run 'GitHubAPISurface|ChatwootAPISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1
git diff --check
```

Result: pass.

## Completed full handoff gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. `make verify` completed `go test -timeout 20m ./...`, `go build ./cmd/pm`, connector docs validation, smoke test, `golangci-lint`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
