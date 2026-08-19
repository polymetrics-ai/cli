# Verification: Issue #4305

## Required checks

- [x] Targeted engine, commandrunner, and installed CLI tests with -timeout 20m.
- [x] Existing scalar, form, SCIM, binary, and specialized GitHub focused regressions.
- [x] go vet ./...
- [x] go build ./cmd/pm
- [x] go run ./cmd/connectorgen validate internal/connectors/defs
- [x] go run ./cmd/connectorgen surface-sync --check
- [x] Generated help/manual/schema checks and applicable website parity check.
- [x] Completion-tracked make connector-boundary.
- [x] make verify.
- [x] git diff --check.
- [x] Code-review evidence with actionable findings resolved.

## Results

### Focused TDD and behavior checks

- `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO|WriteActionOptionalRecordQueryIsOmittedOrPreservedAtProvider|OperationDirectWriteStructuredRESTBody)' -count=1` — pass.
- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen ./internal/cli -run 'Test(OperationDirectWriteStructuredRESTBody|OperationStructuredJSONBodyPreflight|BuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody|GitHubUserDraftCommandBuildsFixedGraphQLMutation|Validate_CLISurfaceDirectWriteStructuredRESTBody|StructuredRESTBodyCommandHelpAndManual|ConnectorsManualDocumentsConnectorArchitectureAndGithubExamples)' -count=1` — pass.
- Full packages passed with `-timeout 20m`: `./internal/connectors/engine`, `./internal/connectors/commandrunner`, `./cmd/connectorgen`, and `./internal/cli`.
- The structured REST fixture proves exact provider request materialization, path/query/body isolation, action isolation, nested approval mutation rejection, and zero transport calls for malformed, missing, unknown, wrong-type, over-limit, and metadata-override inputs. The write-query fixture proves only the exact object-form `record.*` declaration can omit an absent value; supplied values reach the provider and required/undeclared/wrong-source/malformed inputs reject pre-I/O.

### Generated command surface and docs

- Built the actual binary: `go build -o ./pm ./cmd/pm`.
- Runtime help/manual checks passed: `./pm help connectors`, `./pm connectors`, and `./pm docs generate --dir docs/cli --connectors-dir docs/connectors`. A second generation preserved the `docs/cli/connectors.md` content hash.
- `POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=help_connectors,bare_connectors_manual,json_connectors_manual,connectors_help_github_known_legacy_namespace_help_intercept go test -timeout 20m ./internal/cli -run TestGoldenTranscripts -count=1` passed after regenerating the complete affected manual fixture set. `go test -timeout 20m ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1` passed.

### Repository gates

- `go vet ./...` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — pass: 552 connectors, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` — pass: 552 scanned, 0 changes.
- `make connector-boundary` — pass: clean whole-tree report, 294 files checked and 552 connectors loaded; only pre-existing dated exceptions.
- Final `make verify` — pass. It includes `gofmt`, tidy check, `go vet ./...`, `go test -timeout 20m ./...`, `pm` build/docs/smoke, lint (0 issues), agent contract, generator validation/sync, GitHub artifacts, connector boundary, connector canon, pinned dependencies, and release target parity.
- `git diff --check` — pass.

### Delivery note

The explicit Firstmate correction defers the no-mistakes, push, and PR stages. They were not invoked in this worktree.
