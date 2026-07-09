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

Targeted checks passed:

```bash
gofmt -w internal/cli/cli.go internal/cli/cli_test.go internal/connectors/guide.go
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
go test ./internal/connectors/bundleregistry -run CLISurface -count=1
go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm gen:website-data
cd website && pnpm test:unit -- connector-data
git diff --check
```

CLI parity checks passed:

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors
go run ./cmd/pm connectors inspect jira --json
go run ./cmd/pm jira --help
go run ./cmd/pm jira
go run ./cmd/pm --json jira --help
go run ./cmd/pm help jira
```

Connector validation result:

```json
{
  "findings": [],
  "warnings": [],
  "connectors_checked": 547
}
```

Full gates passed before #105 handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

Final connector validation result:

```text
connectorgen validate: 547 connector(s) checked, 0 findings
```

Review-fix checks after Copilot backup comments also passed:

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill -count=1
go run ./cmd/pm docs validate --connectors-dir docs/connectors
go run ./cmd/pm --json help jira
go run ./cmd/pm jira
go run ./cmd/pm version
rg -n "site scope" docs/connectors/jira/MANUAL.md docs/connectors/jira/SKILL.md
go test ./internal/cli -run TestJiraConnectorCommandSurfaceHelp -count=1
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

Note: two early full-suite attempts hit `internal/connectors/certify` timeout/timing flakes under local load. Focused certify reruns passed, `go test -timeout 20m ./...` passed, and the required `go test ./...` subsequently passed before `make verify`.

## Exemptions

- No credentialed Jira checks: help/docs metadata-only slice.
- No reverse ETL execution: writes remain unimplemented.
