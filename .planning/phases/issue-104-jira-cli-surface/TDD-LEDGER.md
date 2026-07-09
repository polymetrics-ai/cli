# TDD Ledger: Issue #104 Jira CLI Surface Metadata

## Preflight

- `scripts/gsd prompt plan-phase issue-104-jira-cli-surface --skip-research`: generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-104-jira-cli-surface --dry-run`: failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active.

## Planned Red

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Expected first failure: `Jira CLISurface is nil; defs.FS must embed cli_surface.json`.

## Red Evidence

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Result: failed as expected.

```text
--- FAIL: TestBundleLoadEmbeddedJiraCLISurface (0.00s)
    bundle_test.go:933: Jira CLISurface is nil; defs.FS must embed cli_surface.json
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.336s
FAIL
```

## Green Evidence

```bash
python3 -m json.tool internal/connectors/defs/jira/cli_surface.json >/dev/null
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./cmd/connectorgen ./internal/connectors/engine -count=1
go test ./internal/connectors/conformance -run 'TestConformance/jira' -count=1
```

Results:

- JSON parse: passed.
- `TestBundleLoadEmbeddedJiraCLISurface`: passed.
- Engine CLI-surface focused tests: passed.
- Connectorgen CLI-surface focused tests: passed.
- Connector definition validation: passed (`connectors_checked: 547`, no findings/warnings).
- `cmd/connectorgen` + `internal/connectors/engine` package tests: passed.
- Jira conformance focused test: passed.

Full verification passed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Final connector validation: `connectorgen validate: 547 connector(s) checked, 0 findings`.

## Refactor Evidence

- Kept `cli_surface.json` metadata-only: no `writes.json`, no operation executor, no generic raw API path.
- Implemented commands only target existing Jira streams and exact `api_surface.json` covered endpoints.
- Future writes and admin/destructive commands are classified as planned or blocked, with approval/risk notes where relevant.

## Notes

- The first production edit must be the failing test, not `cli_surface.json`.
- Implementation remains metadata-only; no Jira credentials or live API checks are required.
