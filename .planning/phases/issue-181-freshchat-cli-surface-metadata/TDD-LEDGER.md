# TDD Ledger — Issue #181 Freshchat CLI surface metadata

## Red target

Test name to add before production metadata:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
```

Observed red result before `internal/connectors/defs/freshchat/cli_surface.json` exists:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
```

Result: fail.

```text
--- FAIL: TestBundleLoadEmbeddedFreshchatCLISurface (0.00s)
    bundle_test.go:933: Freshchat CLISurface is nil; defs.FS must embed cli_surface.json
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.350s
FAIL
```

## Green target

After adding `cli_surface.json`:

```bash
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

Observed green result:

```bash
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'
```

Results:

- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface`: pass (`ok polymetrics.ai/internal/connectors/engine 0.343s`).
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass (`547 connector(s) checked, 0 findings`).
- `go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'`: pass (`ok polymetrics.ai/cmd/connectorgen 0.275s`).
- `go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/commandrunner`: pass.
- Full gate: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs` pass. Note: first uncached `go test ./...` hit Go's default 10m timeout in `internal/connectors/certify`; `go test ./internal/connectors/certify -timeout 20m -count=1` passed, `make verify` passed with `go test -timeout 20m ./...`, and exact `go test ./...` passed on rerun.
- Safety validation kept planned/excluded endpoints non-executable: a temporary `api_surface` annotation on planned/excluded commands produced expected `cli_surface_safety` findings, so final metadata only attaches executable API references to commands backed by streams or write actions.

Green assertions:

- Freshchat CLISurface is non-nil.
- Usage is `pm freshchat <command> [flags]`.
- Implemented ETL command `user list` maps to stream `users`.
- Implemented reverse-ETL command `user create` maps to write action `create_user`.
- Full connector defs validation has zero findings.

## Refactor notes

Keep test assertions focused on metadata existence and safety-critical mappings. Avoid snapshotting the entire command file to reduce churn.

## Safety evidence

Validation must reject secret-shaped literals if accidentally introduced into `cli_surface.json`; use `go run ./cmd/connectorgen validate internal/connectors/defs` as the safety gate.
