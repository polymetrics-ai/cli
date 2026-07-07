# TDD Ledger: GitHub Operation Kernel

Issue: #56
Branch: `feat/56-operation-kernel`

## Manual GSD Fallback

The scripted GSD helpers are unavailable in this checkout. Evidence below is
recorded manually following the repository's required red/green/refactor loop.

## Cycles

### Cycle 1: Operation metadata and command references

Red evidence captured before production changes:

```bash
go test ./internal/connectors/engine -run 'TestBundleLoad.*Operation|TestBundleLoadEmbeddedGitHub'
```

Result: failed to compile because `Bundle.Operations` and `CLICommand.Operation` do not exist.

```bash
go test ./internal/connectors/commandrunner -run TestRunImplementedOperationCommandIsFeatureGated
```

Result: failed to compile because `connectors.CommandSurfaceCommand.Operation` does not exist.

```bash
go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceOperation|TestValidate_CLISurfaceUnknownOperation|TestValidate_CLISurfaceRejectsCommandWithStreamAndOperation'
```

Result: failed because `cli_surface.json` rejects `/commands/0/operation` as an unknown additional property.

Expected failure: current code has no operation kernel metadata, no command operation field, and no operation reference validation.

Green evidence after implementation:

```bash
go test ./internal/connectors/engine -run 'TestBundleLoad.*Operation|TestBundleLoadEmbeddedGitHub'
```

Result: pass.

```bash
go test ./internal/connectors/commandrunner
```

Result: pass.

```bash
go test ./cmd/connectorgen -run 'TestValidate_.*Operation|TestValidate_CLISurface'
```

Result: pass.

### Cycle 2: operations.json load-error attribution

Red evidence captured by temporarily removing the `operations.json` attribution
case from `loadErrorFinding`:

```bash
go test ./cmd/connectorgen -run TestValidate_InvalidOperationsJSONFindingNamesOperationsFile
```

Result: failed because malformed `operations.json` was reported as
`metadata.json`.

Green evidence after restoring the attribution case:

```bash
go test ./cmd/connectorgen -run TestValidate_InvalidOperationsJSONFindingNamesOperationsFile
```

Result: pass.

### Cycle 3: operation semantic validation

Red evidence:

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadRejectsOperationWithoutMatchingBlock|TestBundleLoadRejectsOperationWithMultipleExecutionBlocks|TestBundleLoadRejectsDuplicateOperationIDs'
```

Result: failed because malformed operations loaded successfully.

Green evidence:

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadRejectsOperationWithoutMatchingBlock|TestBundleLoadRejectsOperationWithMultipleExecutionBlocks|TestBundleLoadRejectsDuplicateOperationIDs|TestBundleLoadEmbeddedGitHubOperations|TestBundleLoadParsesOperations'
```

Result: pass.

Additional semantic coverage added for `rest_write` GET rejection and
`binary_download` positive `max_bytes`.

### Cycle 4: pre-credential operation blocking

Red evidence:

```bash
go test ./internal/cli -run TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution
```

Result: failed by opening the project/app before returning a blocked operation
error.

Green evidence:

```bash
go test ./internal/cli -run 'TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution|TestGitHubCommandSurfaceBlocksReverseETLCommand|TestGitHubCommandSurfaceRunsStreamBackedIssueList|TestGitHubCommandSurfaceRunsDirectReadFile'
```

Result: pass.

### Cycle 5: blocked API-surface operation rows are not executable

Red evidence came from the reviewer finding and existing direct-read fixtures:
blocked `api_surface.operation` rows were accepted as executable direct reads.

Green evidence:

```bash
go test ./cmd/connectorgen -run 'TestValidate_CLISurfaceImplementedDirectRead|TestGitHubAPISurfaceOperationLedgerMetrics'
go test ./internal/connectors/engine
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: pass. GitHub REST API surface now records 101 covered endpoints and
402 blocked operation-ledger rows.

### Cycle 6: operations.json secret scanning

Red evidence:

```bash
go test ./cmd/connectorgen -run TestValidate_OperationsSecretLookingLiteralIsHardFinding
```

Result: failed because `operations.json` was not scanned for secret-shaped
literals.

Green evidence:

```bash
go test ./cmd/connectorgen -run TestValidate_OperationsSecretLookingLiteralIsHardFinding
```

Result: pass.

## Final Focused Gates

```bash
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/conformance
go test ./internal/cli -run 'TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution|TestGitHubCommandSurfaceBlocksReverseETLCommand|TestGitHubCommandSurfaceRunsStreamBackedIssueList|TestGitHubCommandSurfaceRunsDirectReadFile'
go test ./cmd/...
go run ./cmd/connectorgen validate internal/connectors/defs --json
go vet ./...
go build ./cmd/pm
```

Result: pass. `connectorgen validate` checked 547 connectors with zero findings
and zero warnings.
