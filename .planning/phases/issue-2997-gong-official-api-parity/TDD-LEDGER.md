# TDD Ledger — Gong Official API Parity Completion (#2997)

## Planned red

Targeted connector-owned red test before production bundle edits:

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Expected current failure after updating test expectations first:

- `endpoints = 67, want 69`, or
- method counts missing `GET /v2/targets` and `POST /v2/targets/{targetId}/assignments`.

## Planned green

After connector-local implementation:

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/gong
go test ./internal/connectors/conformance -run 'TestConformance/gong' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

## Red evidence

### Red — 2026-07-30

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Result: failed as expected against the pre-edit Gong bundle.

```text
--- FAIL: TestGongAPISurfaceOperationLedger (0.00s)
    gong_api_surface_test.go:72: endpoints = 67, want 69
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.345s
FAIL
```

## Green evidence

### Green — 2026-07-30

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/cmd/connectorgen	0.310s
```

```bash
go test ./cmd/connectorgen -count=1
```

Result: pass after updating the connector-owned full-surface test for 69 official operations, 27 writes, 30 direct reads, and the Targets commands.

```text
ok  	polymetrics.ai/cmd/connectorgen	12.164s
```

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass. The issue checklist's narrower `internal/connectors/defs/gong` path is not accepted by the current validator because it treats child directories such as `schemas/` and `fixtures/` as bundles; the valid root-level invocation checked Gong with the rest of the defs tree.

```text
connectorgen validate: 548 connector(s) checked, 0 findings
```

```bash
go test ./internal/connectors/conformance -run 'TestConformance/gong' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/conformance	2.449s
```

```bash
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/cli	192.036s
```

Additional gates passed: `go vet ./internal/connectors/... ./internal/cli/...`, `go build ./cmd/pm`, `make connector-boundary`, and `git diff --check`.

## Refactor notes

- Kept implementation connector-local: Gong defs, connector-owned tests, GSD artifacts, and command metadata only.
- Expanded `operations.json` to 69 rows, one per current official OpenAPI operation. `stream_etl` rows use the required `composite` execution block; direct reads/writes use `rest`.
- Added Targets parity: `targets list` bounded direct read and `upload_target_assignments` typed multipart reverse-ETL action with `confirm: destructive` and typed-confirmation operation policy.
- Left certification at 0; no live credentials, provider calls, or writes were run.
