# Verification: Gorgias CLI Surface Metadata

Date: 2026-07-09 UTC

## Planned focused commands

```bash
jq empty internal/connectors/defs/gorgias/api_surface.json internal/connectors/defs/gorgias/cli_surface.json
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
git diff --check
```

## Planned broader commands

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Results

Pending.
