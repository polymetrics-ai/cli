# Verification: Issue #106 Jira Stream Runner

## Targeted checks

```bash
go test ./internal/cli -run 'TestJiraCommandSurfaceRunsStreamBacked' -count=1
go test ./internal/connectors/commandrunner -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm test:unit -- connector-data
```

## CLI parity checks

```bash
go run ./cmd/pm jira --help
go run ./cmd/pm --json jira --help
go run ./cmd/pm help jira
go run ./cmd/pm connectors inspect jira --json
```

## Full gates before handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

## Results

Pending.
