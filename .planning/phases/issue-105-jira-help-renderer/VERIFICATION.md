# Verification: Issue #105 Jira Help Renderer / Docs

## Targeted Commands

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
go test ./internal/connectors/bundleregistry -run CLISurface -count=1
go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
cd website && pnpm test:unit -- connector-data
```

## Generated Artifacts / Idempotency

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm gen:website-data
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
go run ./cmd/pm jira --help
go run ./cmd/pm jira
rg -n "COMMAND SURFACE|issue list|pm jira" docs/connectors/jira website/data website/lib website/content/docs docs/cli
```

## Results

Pending.

## Exemptions

- No credentialed Jira checks: help/docs metadata-only slice.
- No reverse ETL execution: writes remain unimplemented.
