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

Pending.

## Exemptions

- No credentialed Jira checks: metadata-only slice.
- No reverse ETL execution: writes remain unimplemented.
- No website/docs generator unless this slice changes generated metadata consumers.
