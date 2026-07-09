# TDD Ledger — issue #116 Monday GraphQL/direct-read engine

Manual GSD programming-loop fallback is active.

| Step | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./internal/connectors/engine -run 'TestDirectReadGraphQL' -count=1` fails to compile because `DirectReadRequest` has no `Operation` field yet. | Captured |
| Green | Added `DirectReadRequest.Operation`, `graphql_json` output policy, and fixed `graphql_query` execution via POST with variables from declared flags. Targeted engine/runner tests pass. | Captured |
| Refactor | `go test ./internal/connectors/engine -run 'TestDirectRead' -count=1`, `go test ./internal/connectors/commandrunner -run 'TestRun.*DirectRead' -count=1`, and connectorgen validation pass. | Captured |
