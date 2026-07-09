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

Expected initial failure: current Freshchat api surface has no `operation_ledger_version` and still uses legacy `excluded` rows.

## Green target

- Update `api_surface.json` to ledger mode and replace the three legacy excluded rows with blocked operation rows.
- Preserve all existing stream/write coverage references.
- `go run ./cmd/connectorgen validate internal/connectors/defs` reports zero findings.

## Verification commands

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger
go test ./cmd/connectorgen -run 'TestValidate_APISurfaceOperationLedger|TestValidate_CLISurface'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full parent/subissue gates before handoff when practical:

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```
