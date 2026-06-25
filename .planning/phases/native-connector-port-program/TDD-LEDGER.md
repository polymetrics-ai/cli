# TDD Ledger

Phase: native-connector-port-program

Record failing test evidence before production code for every behavior-adding task.

## Red Tests

- `go test ./internal/connectors -run TestNativePort`
  - Fails because `NativePortPlans`, `NativePortPlanForDefinition`, and port family constants are undefined.
- `go test ./internal/cli -run TestConnectorPortPlan`
  - Fails because `pm connectors port-plan` returns invalid usage.

## Behavior Tasks

- Implement native port plan generation for every catalog connector.
- Classify database CDC source plans for Postgres, MySQL, MongoDB, SQL Server, Oracle, and related database connectors.
- Add CLI and manual rendering for native port plans.

## Green Tests

- `go test ./internal/connectors -run TestNativePort` passed.
- `go test ./internal/cli -run TestConnectorPortPlan` passed.
- `go test ./...` passed.
- `node /Users/karthiksivadas/.codex/skills/gsd-programming-loop/scripts/tdd-gate.mjs --phase native-connector-port-program` passed.
