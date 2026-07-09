# Verification — Issue #80 Linear all-ops update

## Focused checks

```bash
go test ./internal/connectors/engine -run 'TestLinear.*Operation|TestLinearMutationOperationsModeledAsTypedWrites|TestWriteGraphQLVariableSourcePreservesArraysAndObjects|TestLinearWriteActionUsesFixedGraphQLMutation' -count=1
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

git diff --check
# pass
```

## Coverage snapshot

- `api_surface.json` rows: 466
- covered rows: 465
- blocked rows: 1 (`/graphql (raw arbitrary query or mutation)`)
- write actions: 321 fixed GraphQL reverse-ETL actions
- streams: 144 fixed GraphQL streams

No credentialed Linear checks or live Linear writes were run.
