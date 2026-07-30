# TDD Ledger: Issue 2985 Provider Search/Query Foundation

## Manual GSD fallback

`/gsd-programming-loop` / `scripts/gsd prompt programming-loop ...` is unavailable in this checkout. Red/green/refactor evidence is recorded manually.

## Planned cycles

### Cycle 1: capability metadata and validation

Red commands before production edits:

```bash
go test ./cmd/connectorgen -run 'TestValidate_Provider(Search|Query)'
go test ./internal/connectors/engine -run 'TestBundleLoad.*Provider'
```

Expected red: tests fail to compile or fail because `provider_search` / `provider_query` fields and operation kinds do not exist.

Red evidence captured:

```bash
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/cli -run 'TestValidate_Provider|TestValidate_CLISurfaceProvider|TestBundleLoad.*Provider|TestRunImplementedProvider|TestDefinitionProviderOperationsJSONShape|TestGuideSeparates|TestQueryHelpDocuments|TestConnectorsManualDocuments'
```

Result: failed as expected. Connectors and engine tests failed to compile on missing `ProviderSearch`, `ProviderQuery`, `ProviderOperations`, and provider operation structs; connectorgen rejected provider capability fields as unknown metadata properties; CLI help tests missed provider/search query separation text.

Green target:

- loader accepts provider operation contracts;
- validator rejects metadata-only enablement and unsafe raw request fields;
- targeted tests pass.

Green evidence:

```bash
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/cli -run 'TestValidate_Provider|TestValidate_CLISurfaceProvider|TestBundleLoad.*Provider|TestRunImplementedProvider|TestDefinitionProviderOperationsJSONShape|TestGuideSeparates|TestQueryHelpDocuments|TestConnectorsManualDocuments|TestBareCommandShowsManualInsteadOfUsageError|TestValidate_APISurfaceOperationLedgerValidRowsPassCleanly|TestEveryRegisteredConnectorHasGuideManualAndSkill'
```

Result: pass after adding metadata/schema/types for `provider_search` / `provider_query`, provider operation bounds/pagination/schema/fixture validation, and command surface validation for matching provider operation intents.

### Cycle 2: inspect/help separation

Red commands before production edits:

```bash
go test ./internal/connectors -run 'Test.*Provider.*Manual|Test.*Capabilities'
go test ./internal/cli -run 'Test.*Connectors.*Provider|Test.*Query.*Warehouse'
```

Expected red: rendered docs/manuals do not mention provider search/query separately from warehouse query.

Green target:

- connector manual capability sections separate warehouse query from provider search/query;
- CLI help/docs mention ETL streams, direct reads, provider search/query, reverse ETL, and `pm query` local warehouse query distinctly;
- existing `pm query` behavior remains unchanged.

Green evidence:

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$tmpdir/connectors"
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts
go test ./internal/cli -run 'TestQueryHelpDocuments|TestConnectorsManualDocuments|TestBareCommandShowsManualInsteadOfUsageError|TestGoldenDocsGenerateMatchesTrackedCLIManuals'
```

Result: pass. CLI manuals, tracked `docs/cli`, and golden transcripts now describe local warehouse `pm query` separately from provider API search/query metadata.

### Cycle 3: command surface fail-closed behavior

Red commands before production edits:

```bash
go test ./internal/connectors/commandrunner -run 'Test.*Provider.*Unsupported'
go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceProvider'
```

Expected red: provider search/query command intents are rejected as unknown or not classified.

Green target:

- implemented provider commands require matching typed operations;
- commandrunner returns a blocked/unsupported error because no provider executor exists yet.

Green evidence:

```bash
go test ./internal/connectors/commandrunner -run TestRunImplementedProviderSearchCommandIsUnsupportedUntilExecutorLands
go test ./cmd/connectorgen -run 'TestValidate_ProviderSearchOperationReferencePasses|TestValidate_CLISurfaceProviderOperationKindMismatchIsHardFinding'
```

Result: pass. Runtime command dispatch remains fail-closed for provider operation commands until an executor lands.

## Final verification checklist

- [x] gofmt on touched Go files
- [x] targeted Go tests for connectorgen, engine, commandrunner, connectors, internal/cli
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] issue acceptance checks, including docs/help grep and no raw escape-hatch validation
- [ ] after commit, fetch/rebase onto `origin/feat/connector-wave-01` when available

Additional result:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: pass, 548 connectors, zero findings and zero warnings.

Website note: `cd website && npm run typecheck` could not run because dependencies are not installed in this worktree (`tsc: command not found`); website source/generator/types were updated and no package installation was performed.
