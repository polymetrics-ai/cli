# TDD Ledger — Issue #184 Freshchat operation ledger

## Red target

Add `cmd/connectorgen/freshchat_api_surface_test.go` with `TestFreshchatAPISurfaceOperationLedger`:

- loads `internal/connectors/defs/freshchat/api_surface.json`;
- requires `operation_ledger_version == 1`;
- requires exactly 34 endpoints;
- requires exactly 18 `covered_by.stream` rows;
- requires exactly 13 `covered_by.write` rows;
- requires exactly 3 blocked `operation` rows;
- requires blocked operation model counts: `direct_read: 1`, `disallowed: 2`;
- rejects any legacy `excluded` row.

Observed initial failure:

```bash
gofmt -w cmd/connectorgen/freshchat_api_surface_test.go
go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger
```

Result:

```text
--- FAIL: TestFreshchatAPISurfaceOperationLedger (0.00s)
    freshchat_api_surface_test.go:30: operation_ledger_version = 0, want 1
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.384s
```

This matches the expected failure: current Freshchat api surface has no `operation_ledger_version` and still uses legacy `excluded` rows.

## Green result

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger
go test ./cmd/connectorgen -run 'TestValidate_APISurfaceOperationLedger|TestValidate_CLISurface'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- `go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger`: pass.
- `go test ./cmd/connectorgen -run 'TestValidate_APISurfaceOperationLedger|TestValidate_CLISurface'`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.

Green changes:

- Updated `api_surface.json` to ledger mode and replaced the three legacy excluded rows with blocked operation rows.
- Preserved all existing stream/write coverage references.
- Verified all Freshchat 34 endpoints are accounted: 18 streams, 13 writes, 3 blocked operations.

## Verification commands

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger
go test ./cmd/connectorgen -run 'TestValidate_APISurfaceOperationLedger|TestValidate_CLISurface'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full parent/subissue gates before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass. `make verify` included docs validation, smoke, lint, and connectorgen validation; final standalone connectorgen validation reported `547 connector(s) checked, 0 findings`.
