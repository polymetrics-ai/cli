# Verification: Zendesk CLI Surface Metadata

## Planned commands

```bash
test -d internal/connectors/defs/zendesk
go test ./internal/connectors/engine -run 'CLISurface|Definition|Zendesk' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI parity checks

Metadata-only exemptions for #157:

- Runtime help (`pm help zendesk`, `pm zendesk --help`) — not applicable until #158 implements/updates rendering.
- Bare namespace command (`pm zendesk`) — not applicable until #158/runtime dispatcher.
- `docs/cli/**` — not applicable until #158 unless #157 changes runtime docs.
- `website/**` — not applicable until #158 unless #157 changes generated website metadata.
- Generated help/manual artifacts — not applicable until #158.

## Results

Pending implementation.
