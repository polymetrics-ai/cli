# Verification: Chatwoot CLI Surface Metadata

## Targeted gates

```bash
python3 .planning/phases/issue-149-chatwoot-cli-surface-metadata/traces/verify-official-surface-count.py
jq empty internal/connectors/defs/chatwoot/api_surface.json internal/connectors/defs/chatwoot/cli_surface.json
go test ./cmd/connectorgen -run CLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
```

Note: `go run ./cmd/connectorgen validate internal/connectors/defs/chatwoot` is not used because the current validator expects the defs root; when pointed at a connector subdirectory it interprets nested `fixtures/` and `schemas/` as connector directories.

## Broader issue gate

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Final parent gate (not required before #149 sub-slice handoff unless parent is being marked ready)

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity status

- Runtime help: deferred to #150; no renderer behavior changed in #149.
- Bare namespace behavior: deferred to #150.
- `docs/cli/**`: deferred to #150.
- `website/**`: deferred to #150.
- Generated help/manual artifacts: deferred to #150.
- Connector `docs.md`: in scope for #149.
