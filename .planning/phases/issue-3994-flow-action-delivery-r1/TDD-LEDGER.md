# TDD Ledger — Issue #3994

## Red

- 2026-08-15: `go test -timeout 20m ./internal/app -run 'TestExecuteAuthorizedFlowAction'` failed before implementation because `App.ExecuteAuthorizedFlowAction`, the flow action request/result types, and receipt API did not exist. This made the missing public connector-action boundary observable at compile time.
- 2026-08-15: added a production-composition test where an approved scope's field mapping changes. The target records every validation, write, and read-back operation; the required red behaviour was zero new target events, zero writes, zero receipts, and no checkpoint.

## Green

- 2026-08-15: `App.ExecuteAuthorizedFlowAction` resolves the typed connector and credential, re-derives/requires the durable #4132 authorization scope, maps selected records, validates and re-previews a destructive target, writes with durable evidence, reads the configured target stream, then persists an opaque receipt.
- 2026-08-15: `connectorFlowActionRunner` is assembled by `flowRun`, reads the selected warehouse source through Engine, and checkpoints only after the receipt-returning action result. It accepts a manifest `authorization_reference` or a `--authorization` override.
- 2026-08-15: `TestConnectorFlowActionRunnerScopeDriftStopsBeforeTargetRequest` proves a changed mapping performs zero target validation/write/read-back calls, zero receipts, and no checkpoint.
- 2026-08-15: `TestExecuteAuthorizedFlowActionAcknowledgesReadsBackThenRecordsReceipt` proves a destructive target receives and verifies real durable authorization evidence, and receipt count remains zero while connector `Write` and `Read` run; `TestExecuteAuthorizedFlowActionReadBackFailurePersistsNoReceipt` proves a failed read-back leaves no receipt after an acknowledged write.

## Refactor

- 2026-08-15: kept the existing HTTP runner only as a legacy flow-package fixture and documented that `flowRun` exclusively uses the connector-backed runner. Restored the manifest's documented `upsert` default in the runner.

## Verification log

- 2026-08-15: `scripts/gsd doctor` passed; `scripts/gsd sources` resolved every required lifecycle command; `go run ./cmd/agentcontractgen check` passed.
- 2026-08-15: rebased on `origin/integration/4015-mvp-flat-r1`; verified `internal/app/authorization.go` and `internal/connectors/database/managed_target_delivery_ledger.go` exist before design.
- 2026-08-15: green targeted tests: `go test -timeout 20m ./internal/flow/... ./internal/app/... ./internal/cli -run 'Test(ExecuteAuthorizedFlowAction|EngineAction|FlowRun|ConnectorFlowActionRunner)'`.
- 2026-08-15: green scoped suites: `go test -timeout 20m ./internal/flow/...`, `go test -timeout 20m ./internal/app/...`, and `go test -timeout 20m ./internal/cli`.
