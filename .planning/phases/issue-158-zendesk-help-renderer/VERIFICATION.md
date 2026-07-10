# Verification: Zendesk Help Renderer

## Planned checks

```bash
go test ./internal/cli -run 'Zendesk.*Help|Connector.*Help' -count=1
./pm help zendesk
./pm zendesk
./pm zendesk read list-tickets --help
./pm connectors inspect zendesk --json
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
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
go test ./internal/cli -run 'Zendesk.*Help' -count=1
go test ./internal/connectors/conformance -run 'TestConformance/zendesk$' -count=1
./pm help zendesk
./pm zendesk
./pm zendesk read list-tickets --help
./pm connectors inspect zendesk --json
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Notes:

- First full `go test ./...` failed in `internal/connectors/conformance` because Zendesk streams lacked a mandatory first-stream fixture. Added a synthetic `list_activities` replay page and reran the gates successfully.
- `make verify` included `go mod tidy`, `go vet`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validation, smoke, golangci-lint, and `connectorgen validate`.

## Safety verification

- Help checks are non-credentialed.
- No live Zendesk requests are run.
- No reverse ETL execution is run.
