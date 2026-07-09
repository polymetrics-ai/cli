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

Completed targeted checks:

```bash
jq empty internal/connectors/defs/zendesk/metadata.json internal/connectors/defs/zendesk/spec.json internal/connectors/defs/zendesk/streams.json internal/connectors/defs/zendesk/api_surface.json internal/connectors/defs/zendesk/cli_surface.json
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedZendeskCLISurface -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface' -count=1
go test ./internal/connectors/engine -run 'CLISurface|Zendesk' -count=1
go test ./internal/connectors/bundleregistry -count=1
go test ./internal/connectors/conformance -run 'TestConformance/zendesk$' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/bundleregistry ./internal/connectors/conformance -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go run ./cmd/pm help connectors
go run ./cmd/pm connectors inspect zendesk --json
```

Results:

- JSON parse checks passed.
- Focused Zendesk embedded metadata test passed.
- Focused CLI/API surface validator tests passed.
- Bundleregistry tests passed after updating the expected embedded bundle count from 547 to 548.
- Zendesk conformance passed; zero streams means fixture-required checks are vacuously satisfied.
- Focused package test batch passed.
- `connectorgen validate` passed: 548 connectors checked, 0 findings, 0 warnings.
- `pm help connectors` rendered before using `pm connectors inspect`.
- `pm connectors inspect zendesk --json` returned the redacted metadata/spec fields and did not read credentials.

Broad handoff gates completed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- `gofmt -w cmd internal` completed.
- `go vet ./...` passed.
- `go test ./...` passed. Note: an earlier uncached run hit the default 10-minute package timeout in `internal/connectors/certify`; after reducing runtime-embedded Zendesk CLI surface metadata and warming the package with `go test -timeout 20m`, the required command completed successfully.
- `go build ./cmd/pm` passed.
- `make verify` passed, including docs validation, smoke, lint, and connector validation. The first `make verify` run exposed stale generated connector docs count; fixed by updating only the connector catalog artifacts and new `docs/connectors/zendesk/` files from a temp docs generation, avoiding a broad generated-doc rewrite.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed: 548 connector(s) checked, 0 findings.
