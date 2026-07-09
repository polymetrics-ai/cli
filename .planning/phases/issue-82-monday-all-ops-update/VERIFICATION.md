# Verification — issue #82 Monday all-ops update

```bash
go test ./cmd/connectorgen -run 'TestMondayFullSurfaceAllOpsCovered' -count=1
go test ./internal/connectors/hooks/monday -run 'TestMondayWriteHookBlocksModeledMutations' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./cmd/connectorgen -run 'TestMondayFullSurfaceAllOpsCovered' -count=1` — pass.
- `go test ./internal/connectors/hooks/monday -run 'TestMondayWriteHookBlocksModeledMutations' -count=1` — pass.
- `go test ./cmd/connectorgen -run 'TestMonday' -count=1` — pass.
- `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMonday' -count=1` — pass.
- `go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1` — pass.
- `go test ./internal/connectors/commandrunner -run 'TestRunMonday' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.
- Full gates passed after the all-ops update:
  - `gofmt -w cmd internal`
  - `go vet ./...`
  - `go test ./...`
  - `go build ./cmd/pm`
  - `go run ./cmd/connectorgen validate internal/connectors/defs`
  - `make verify`
