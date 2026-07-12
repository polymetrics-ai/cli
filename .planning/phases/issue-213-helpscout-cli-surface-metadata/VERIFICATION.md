# Verification: Help Scout CLI Surface Metadata

Date: 2026-07-09

## Planned Commands

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go build ./cmd/pm
```

## CLI Help / Docs / Website Parity

Metadata-only slice. If runtime help or docs are changed, run and record:

```bash
pm help connectors
pm connectors
pm connectors inspect help-scout --help
rg -n "help-scout|Help Scout" docs/cli docs/connectors website
```

## Current Results

Pending implementation.
