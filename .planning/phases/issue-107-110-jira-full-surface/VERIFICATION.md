# Verification: Issues #107-#110 Jira Full Surface

## Targeted checks

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedJiraFullSurface|TestJiraFullSurfacePolicy' -count=1
go test ./internal/connectors/commandrunner -run 'DirectRead|Write' -count=1
go test ./internal/cli -run 'JiraConnectorCommandSurface|JiraCommandSurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/jira --json
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

Targeted checks passed:

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedJiraFullSurface|TestBundleLoadDiskJiraAPISurfaceFullCoverage|TestDirectReadRedactsGenericJSONSensitiveFields' -count=1
go test ./internal/connectors/engine -count=1
go test ./internal/connectors/commandrunner -count=1
go test ./internal/cli -run 'TestJiraCommandSurfaceRunsGeneratedDirectRead|TestJiraCommandSurfaceRunsStreamBackedCommands|TestJiraConnectorCommandSurfaceHelp' -count=1
go test ./cmd/connectorgen -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm test:unit -- connector-data
```

CLI parity checks passed without credentialed Jira calls:

```bash
go run ./cmd/pm jira --help
go run ./cmd/pm --json jira --help
go run ./cmd/pm connectors inspect jira --json
```

Full gates passed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

Final connector validation:

```text
connectorgen validate: 547 connector(s) checked, 0 findings
```
