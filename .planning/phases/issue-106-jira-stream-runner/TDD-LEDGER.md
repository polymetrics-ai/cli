# TDD Ledger: Issue #106 Jira Stream Runner

## Preflight

```bash
scripts/gsd prompt plan-phase issue-106-jira-stream-runner --skip-research
scripts/gsd prompt programming-loop init --phase issue-106-jira-stream-runner --dry-run
```

Result: plan prompt generated; programming-loop command unavailable with `unknown GSD command: programming-loop`; manual fallback active.

## Planned red tests

```bash
go test ./internal/cli -run 'TestJiraCommandSurfaceRunsStreamBacked' -count=1
```

Expected first failures:

- `pm jira issue list --jql ...` rejects `--jql` because Jira stream command flags are not declared.
- `pm jira project list --query ...` rejects `--query` because Jira stream command flags are not declared.
- `pm jira user list --query ...` rejects `--query` because Jira stream command flags are not declared.

## Red evidence

```bash
gofmt -w internal/cli/cli_test.go
go test ./internal/cli -run TestJiraCommandSurfaceRunsStreamBackedCommands -count=1
```

Result: failed as expected.

```text
unknown flag --jql for command "issue list"
unknown flag --query for command "project list"
unknown flag --query for command "user list"
```

## Green evidence

```bash
python3 -m json.tool internal/connectors/defs/jira/cli_surface.json >/dev/null
go test ./internal/cli -run TestJiraCommandSurfaceRunsStreamBackedCommands -count=1
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm gen:website-data && pnpm test:unit -- connector-data
go test ./internal/connectors/commandrunner -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: passed. Jira stream-backed commands now declare bounded read-query flags:

- `issue list`: `--jql`, `--fields`, `--expand`
- `project list`: `--query`, `--type-key`, `--category-id`, `--status`
- `user list`: `--query`, `--account-id`, `--username`

All flags map to `query.*` and execute through the existing generic connector runner; no Jira writes or raw API escape hatches were added.

Full verification passed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

Final connector validation: `connectorgen validate: 547 connector(s) checked, 0 findings`.

## Refactor evidence

- Kept runner code unchanged; issue #106 is a metadata/test/docs slice because the generic command runner already supports implemented ETL streams.
- Regenerated only Jira connector docs and website connector data affected by the Jira CLI-surface changes.
