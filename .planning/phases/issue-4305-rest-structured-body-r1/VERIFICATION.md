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

### CI remediation — 2026-08-21

- Red reproduced locally: `go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectWriteStructuredRESTBody' -count=1` failed to compile because `interpolateDeclaredHeader` called the hardened four-argument `resolveExpr` with its obsolete three-argument form.
- Green: restored the explicit `allowControlCharacters=false` policy at that declaration-bound header call. `go test -timeout 20m ./internal/connectors/engine`, `go test -timeout 20m ./internal/connectors/commandrunner`, `go test -timeout 20m ./internal/app`, `go test -timeout 20m ./internal/cli`, `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m -run '^$' ./internal/cli/ ./internal/connectors/certify/`, `go build -trimpath -ldflags='-s -w' ./cmd/pm`, `go vet ./...`, `go run ./cmd/connectorgen boundary . --json`, `go run ./cmd/connectorgen validate internal/connectors/defs`, `go run ./cmd/connectorgen surface-sync --check`, and `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` passed.
- Website CI/data green evidence: regenerated `website/lib/docs.generated.ts` with `npm run gen:website-data`; repeated generation was stable. `pnpm run typecheck` passed, and `pnpm run lint` completed with only the repository's existing 13 warnings.
- This was an assigned CI repair only. No no-mistakes pipeline command, push, PR mutation, or CI re-run was invoked.
