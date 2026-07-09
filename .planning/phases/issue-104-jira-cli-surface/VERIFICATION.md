# Verification: Issue #104 Jira CLI Surface Metadata

## Targeted Commands

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

## Full Commands Before Handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI Help / Docs / Website Parity

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors
go run ./cmd/pm connectors inspect jira --json
rg -n "jira|Jira" docs/cli website internal/connectors/defs/jira
```

## Results

Targeted checks passed:

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

Connector validation result:

```json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 547
}
```

Full gates passed before #104 handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Final connector validation result:

```text
connectorgen validate: 547 connector(s) checked, 0 findings
```

## Exemptions

- No credentialed Jira checks: metadata-only slice.
- No reverse ETL execution: writes remain unimplemented.
- No website/docs generator unless this slice changes generated metadata consumers.
