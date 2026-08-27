# Verification — #4366

## Planned checks

- `go test -timeout 20m ./cmd/connectorgen -run 'Test.*(Composition|SourceProjection|SourceImport|DeclarationAdmission|OperationEvidence)' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -run 'Test.*(Schema|Structured|Deferred|Operation)' -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -run 'Test.*(Preflight|Deferred|Structured)' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go run ./cmd/connectorgen source-import --check <touched connectors>` or the repository-supported equivalent
- `go run ./cmd/connectorgen validate <touched connectors>` and `go run ./cmd/connectorgen surface-sync --check <touched connectors>` or the supported all-connector checks
- generated operation-evidence/declaration-admission/endpoint-ledger checks
- `go build ./cmd/pm`
- isolated-project credential-boundary checks for every newly runnable command; no credential or live call
- `gofmt -w` on changed Go files, `go vet ./...`, relevant direct-PR verification gates, and `git diff --check`

## CLI parity assessment

Pending implementation. No new generic CLI surface is permitted. If generated flags/help change for an already-admitted command, run `pm help <topic>`, `pm <namespace>`, `pm <command> --help`, inspect `docs/cli/**` and `website/**`, and run the generator parity check. Otherwise record the precise no-change evidence.

## Results

Pending red/green execution.
