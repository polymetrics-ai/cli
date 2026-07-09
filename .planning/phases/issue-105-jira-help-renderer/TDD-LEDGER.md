# TDD Ledger: Issue #105 Jira Help Renderer / Docs

## Preflight

- `scripts/gsd prompt plan-phase issue-105-jira-help-renderer --skip-research`: generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-105-jira-help-renderer --dry-run`: failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active.

## Planned Red

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
cd website && pnpm test:unit -- connector-data
```

Expected first failures:

- Runtime CLI does not render connector help for `pm jira --help` or bare `pm jira`.
- Website data route does not yet assert Jira `cliSurface` metadata.

## Red Evidence

Runtime CLI red test after adding tests:

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
```

Result: failed as expected.

```text
--- FAIL: TestJiraConnectorCommandSurfaceHelp/help_flag
    cli_test.go:155: Run([jira --help]) code = 1 stderr = error: help topic "jira" not found
--- FAIL: TestJiraConnectorCommandSurfaceHelp/help_subcommand
    cli_test.go:155: Run([jira help]) code = 1 stderr = error: help topic "jira" not found
--- FAIL: TestBareJiraConnectorCommandShowsHelp
    cli_test.go:181: Run(jira) code = 2 stderr = error: missing connector command path
--- FAIL: TestJiraConnectorCommandSurfaceHelpJSON
    cli_test.go:195: Run(--json jira --help) code = 1 stderr = error: help topic "jira" not found
```

Website red test after adding Jira expectation:

```bash
cd website && npm ci
cd website && pnpm test:unit -- connector-data
```

Result: failed as expected once dependencies were installed from the existing lockfile.

```text
FAIL tests/api/connector-data.test.ts > connector data route > returns Jira CLI surface metadata for docs rendering
AssertionError: expected null to match object ...
Received: null
```

Note: `pnpm install --frozen-lockfile` was attempted first and failed with `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`; no lockfile changes were made. `npm ci` succeeded without modifying tracked files.

## Green Evidence

Runtime CLI implementation passed:

```bash
gofmt -w internal/cli/cli.go internal/cli/cli_test.go internal/connectors/guide.go
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
go run ./cmd/pm jira --help
go run ./cmd/pm jira
go run ./cmd/pm --json jira --help
```

Website/docs implementation passed:

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm gen:website-data
cd website && pnpm test:unit -- connector-data
```

Connector validation passed:

```bash
go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill -count=1
go test ./internal/connectors/bundleregistry -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `pm jira --help`, `pm jira help`, bare `pm jira`, and `--json jira --help` render the Jira connector manual with `COMMAND SURFACE` and exit successfully.
- Website connector data route returns non-null Jira `cliSurface` metadata.
- Connector docs validate with the Jira `COMMAND SURFACE` section retained in generated manual/skill artifacts.
- Connector validation passed with `connectors_checked=547`, `findings=0`, `warnings=0`.

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

Note: two early full-suite attempts hit `internal/connectors/certify` timeout/timing flakes under local load. Focused certify reruns passed, `go test -timeout 20m ./...` passed, and the required `go test ./...` subsequently passed before `make verify`.

## Review Fix Evidence

Copilot backup review was requested after CodeRabbit reported a rate-limit window on the #105 head. Four actionable comments were dispositioned and fixed:

- Threaded `jsonOut` through `runHelp`, so `pm --json help jira` returns the same `CommandManual` JSON envelope as `pm --json jira --help`.
- Added CLI test coverage for `pm --json help jira`.
- Reformatted command-surface flag metadata so generated docs render `Use a saved Jira connector credential and site scope. (maps_to=connection)` instead of `site scope.: maps_to=connection`.
- Removed the pre-switch `isConnectorCommandSurface` registry load; bare connector commands now fall through to `runMaybeConnectorCommand`, which renders connector help without adding registry startup cost to built-in bare commands.

Review-fix verification passed:

```bash
gofmt -w internal/cli/cli.go internal/cli/cli_test.go internal/connectors/guide.go
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill -count=1
go run ./cmd/pm --json help jira
go run ./cmd/pm jira
go run ./cmd/pm version
rg -n "site scope" docs/connectors/jira/MANUAL.md docs/connectors/jira/SKILL.md
```

Full review-fix verification passed:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

## Refactor Evidence

- Added generic connector command-surface manual routing instead of Jira-specific special cases.
- Bare connector command namespaces with no command path now render contextual connector help successfully.
- Kept unknown connectors and connectors without command surfaces returning usage/help errors.
- Removed trailing empty descriptions from rendered stream/write labels via a shared helper.
