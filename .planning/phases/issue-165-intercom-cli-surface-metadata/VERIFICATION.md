# Verification: Intercom CLI Surface Metadata

## Focused Gates

```bash
jq empty internal/connectors/defs/intercom/api_surface.json internal/connectors/defs/intercom/cli_surface.json .planning/phases/issue-165-intercom-cli-surface-metadata/RUN-STATE.json
go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1
go test ./internal/connectors/engine -run 'Intercom|CLISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1
```

## Broader Gates Before Handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI Help / Docs / Website Parity

- Runtime help: deferred to #166; #165 adds metadata only.
- Bare namespace behavior: deferred to #166.
- `pm <command> --help`: deferred to #166.
- `docs/cli/**`: deferred to #166 unless generated metadata checks require an update.
- `website/**`: deferred to #166 unless generated metadata checks require an update.
- Generated help/manual artifacts: deferred to #166.

## Results

Pending.
