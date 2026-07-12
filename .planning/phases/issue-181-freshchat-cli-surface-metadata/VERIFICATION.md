# Verification — Issue #181 Freshchat CLI surface metadata

## Focused gates run

```bash
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'
```

Results:

- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.
- `go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'`: pass.
- `go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/commandrunner`: pass.

## Optional no-credential smoke checks

Only after `go build ./cmd/pm` and without credentials:

```bash
go build ./cmd/pm
./pm help connectors
./pm connectors --help
./pm freshchat --help || true
```

Results:

- `go build ./cmd/pm`: pass.
- `./pm help connectors`: pass; output begins `NAME`.
- `./pm connectors --help`: pass; output begins `NAME`.
- `./pm freshchat --help || true`: existing help topic is not registered; stderr begins `error: help topic "freshchat" not found`. This is deferred to #182 help renderer/docs parity.
- `./pm connectors inspect freshchat --json`: pass; confirms connector inspect path is credential-free, but current inspect output does not expose command-surface metadata.

No credentialed Freshchat command and no reverse ETL execution were run.

## Full parent gate before handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- `gofmt -w cmd internal`: pass.
- `go vet ./...`: pass.
- `go test ./...`: pass on rerun. The first uncached run hit Go's default 10m timeout in `internal/connectors/certify`; `go test ./internal/connectors/certify -timeout 20m -count=1` then passed, and a subsequent exact `go test ./...` passed with cache coverage.
- `go build ./cmd/pm`: pass.
- `make verify`: pass, including `go test -timeout 20m ./...`, docs validation, smoke, lint, and connectorgen validation.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.

Notes:

- A temporary attempt to attach `api_surface` references to planned/excluded endpoints failed `connectorgen validate` with `cli_surface_safety` findings, as intended. The final metadata leaves executable `api_surface` references only on implemented/partial commands backed by streams or writes; final validation passes with zero findings.
