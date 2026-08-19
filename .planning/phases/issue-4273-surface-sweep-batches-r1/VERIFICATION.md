# Verification — issue #4273 connector surface sweep batch 1

## Planned checks

- `go test -timeout 20m ./cmd/connectorgen`
- `go run ./cmd/connectorgen batch plan ... --size 20`
- `go run ./cmd/connectorgen batch materialize ...`
- `go run ./cmd/connectorgen batch gate ...`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/connectorgen surface-reconcile --check`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`
- `go run ./cmd/connectorgen certification-candidates --connector <name>` for every included bundle
- `make connector-boundary`, `make connector-runtime-preflight`, `go vet ./...`, `go build ./cmd/pm`, individual `make verify` gates, then `make verify`

## Safety boundaries

- No credentialed provider API operation, reverse execution, or live certification is part of this batch.
- Artifact retrieval is bounded to the existing materializer's cited public URL validation, not browser collection.
- Any failed materialization/gate result is reflected in the ledger and either skipped or reverted from the branch before PR creation.
