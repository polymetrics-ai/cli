# TDD Ledger — Issue #80 Linear all-ops update

| Step | Evidence | Result |
|---|---|---|
| Plan | Created after reading refreshed `PI_CONNECTOR_PROMPT.md`; manual-GSD fallback recorded. | done |
| Red | `go test ./internal/connectors/engine -run TestLinearMutationOperationsModeledAsTypedWrites -count=1` failed with 317 mutation rows neither typed writes nor exact hard blocks. | done |
| Green | Added GraphQL source-variable support and generated fixed-document Linear coverage: 321 write actions and 144 query streams; API surface now has 465 covered rows and 1 blocked raw-GraphQL row. | done |
| Verify | Focused engine/CLI/conformance/connectorgen checks passed; parent gates pending. | in_progress |

No credentialed Linear checks or live Linear writes were run.
