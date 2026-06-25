# TDD Ledger

Phase: native-runtime-capability-matrix

Record failing test evidence before production code for every behavior-adding task.

## Red Tests

- `go test ./internal/connectors -run TestConnectorCatalog`
  - Fails because `ConnectorDefinition.RuntimeCapabilities` is not defined.
- `go test ./internal/cli -run TestConnectorCatalog`
  - Fails because catalog JSON lacks `runtime_capabilities`.
  - Fails because catalog-only manuals lack `RUNTIME CAPABILITIES`.

## Behavior Tasks

- Add generated runtime capabilities to every connector definition.
- Render runtime capabilities in catalog JSON, manuals, skills, and generated docs.

## Green Tests

- `go test ./internal/connectors -run TestConnectorCatalog`
  - Passes after adding `RuntimeCapabilities` and catalog fallback generation.
- `go test ./internal/cli -run TestConnectorCatalog`
  - Passes after JSON/manual rendering includes runtime capability data.
- `node /Users/karthiksivadas/.codex/skills/gsd-programming-loop/scripts/tdd-gate.mjs --phase native-runtime-capability-matrix`
  - Passes.
