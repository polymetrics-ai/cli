# TDD Ledger — Issue #80 Linear all-ops update

| Step | Evidence | Result |
|---|---|---|
| Plan | Created after reading refreshed `PI_CONNECTOR_PROMPT.md`; manual-GSD fallback recorded. | done |
| Red | First all-ops RED: `TestLinearMutationOperationsModeledAsTypedWrites` failed with 317 mutation rows neither typed writes nor exact hard blocks. | done |
| Red | Prompt-refresh RED: `go test ./internal/connectors/engine -run 'TestLinearOperationLedgerInventoriesGraphQLOperations|TestLinearMutationOperationsModeledAsTypedWrites' -count=1` failed because the ledger had only 144 query rows and 321 mutation rows, below the live schema inventory target (161 query, 370 mutation; official non-deprecated target 156 query, 358 mutation). | done |
| Green | Added 17 missing live query streams, 48 missing non-deprecated mutation write actions, and a blocked deprecated row for `integrationSettingsUpdate`; ledger now has 161 query rows, 370 mutation rows, 530 covered rows, and 2 blocked rows. | done |
| Verify | Focused Linear tests, connectorgen validation, docs validate, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `git diff --check` passed. | done |

No credentialed Linear checks or live Linear writes were run.
