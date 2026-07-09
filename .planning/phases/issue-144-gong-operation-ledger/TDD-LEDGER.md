# TDD Ledger — Gong Operation Ledger (#144)

## Red

Planned failing test:

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Expected red failure against current bundle:

- `operation_ledger_version = 0, want 1`, or
- `endpoints = 10, want 67`, and/or
- legacy `excluded` rows present.

## Green

Target commands after updating `api_surface.json` and docs:

```bash
gofmt -w cmd/connectorgen/gong_api_surface_test.go
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -count=1
```

## Green

### Green — 2026-07-09

```bash
gofmt -w cmd/connectorgen/gong_api_surface_test.go && go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/cmd/connectorgen	0.305s
```

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass.

```text
connectorgen validate: 547 connector(s) checked, 0 findings
```

```bash
go test ./cmd/connectorgen -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/cmd/connectorgen	6.207s
```

```bash
go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/conformance	2.435s
```

```bash
make verify
```

Result: pass. `make verify` runs `go test -timeout 20m ./...`; a separate exact `go test ./...` attempt failed on existing `internal/connectors/certify` default 10m timeout, so the timeout-adjusted whole-repo gate is the passing evidence.

## Refactor

- Kept operation-ledger rows metadata-only.
- No executor paths added in this slice.
- No operation-ledger validation weakened.
- Preserved full-surface safety: blocked rows point to concrete follow-up lanes (#145/#146/#147).

## Results

### Red — 2026-07-09

```bash
gofmt -w cmd/connectorgen/gong_api_surface_test.go && go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Result: fail as expected.

```text
--- FAIL: TestGongAPISurfaceOperationLedger (0.00s)
    gong_api_surface_test.go:30: operation_ledger_version = 0, want 1
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.381s
FAIL
```

## 2026-07-10 follow-on integration note

The operation-ledger assertions were updated from the initial inventory-only slice to the integrated CLI parity state: 51 executable covered endpoints and 16 typed blocked operation rows. The exact 67-operation public spec inventory remains the source of truth.
