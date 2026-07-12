# Verification — Issue #80 Linear all-ops update

## Focused checks

```bash
go test ./internal/connectors/engine -run 'TestLinearOperationLedgerInventoriesGraphQLOperations|TestLinearMutationOperationsModeledAsTypedWrites|TestWriteGraphQLVariableSourcePreservesArraysAndObjects|TestLinearWriteActionUsesFixedGraphQLMutation' -count=1
# pass

go test ./internal/cli -run 'TestLinear' -count=1
# pass

go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1
# pass

go run ./cmd/connectorgen validate internal/connectors/defs --json
# pass: 0 findings, 547 connectors checked

go run ./cmd/pm docs validate --connectors-dir docs/connectors
# pass

npm --prefix website run gen:website-data
# pass

git diff --check
# pass
```

## Parent gates

```bash
go vet ./...
# pass

go test ./...
# pass

go build ./cmd/pm
# pass

./pm docs validate --connectors-dir docs/connectors
# pass

make verify
# pass
```

Note: the combined parent-gate command timed out during the first `make verify` attempt after earlier gates had passed; rerunning `make verify` with a longer timeout passed.

## Coverage snapshot

- `api_surface.json` rows: 532
- live Linear GraphQL root fields inventoried: 531 (161 query + 370 mutation)
- official non-deprecated prompt target: 514 fields (156 query + 358 mutation)
- covered rows: 530
- blocked rows: 2
  - raw arbitrary GraphQL query/mutation execution: disallowed
  - `integrationSettingsUpdate`: deprecated by the live Linear schema, blocked with exact evidence
- write actions: 369 fixed GraphQL reverse-ETL actions
- streams: 161 fixed GraphQL streams

No credentialed Linear checks or live Linear writes were run.
