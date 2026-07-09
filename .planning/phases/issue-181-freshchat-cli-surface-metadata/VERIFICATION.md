# Verification — Issue #181 Freshchat CLI surface metadata

## Planned focused gates

```bash
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'
```

## Optional no-credential smoke checks

Only after `go build ./cmd/pm` and without credentials:

```bash
./pm help connectors
./pm connectors --help
./pm freshchat --help || true
```

Do not run credentialed Freshchat commands or reverse ETL execution.

## Full parent gate before handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full gate may be deferred until parent integration; record exact blockers if not run.
