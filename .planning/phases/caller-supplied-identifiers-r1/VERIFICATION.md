# VERIFICATION — caller-supplied identifier sets

Status: passed locally on 2026-08-06.

## Acceptance checklist

- [x] Closed operation declaration defines a name, shape, wire, minimum, and mandatory maximum.
- [x] Test-only bundle reaches its operation command and none of the commands is an ETL stream.
- [x] All four declared encodings have end-to-end wire assertions.
- [x] Bounds and malformed input fail before a server receives a request and never echo a supplied identifier.
- [x] Explicit blank and absent list flags have distinct tested outcomes.
- [x] Provenance/coverage remains evidence-only; output-policy schema/runtime drift guard still passes.
- [x] Migration conventions document authorship and the nested-batch decision.

## Commands passed

- Focused: `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -run 'TestCallerSuppliedIdentifier|TestRunCallerSuppliedIdentifier|TestRunOperationDirectReadPreservesIdentifierSetPresence|TestSyncBundleDerivesCallerSuppliedIdentifierSetMapping|TestValidate_CallerSuppliedIdentifierSetBundlePasses' -count=1`
- Packages: `go test ./internal/connectors/engine -count=1`, `go test ./internal/connectors/commandrunner -count=1`, `go test ./cmd/connectorgen -count=1`, and `go test ./internal/cli -count=1` (537s).
- Build/static: scoped `go vet`, `go build ./cmd/pm`, `git diff --check`, `go test ./internal/connectors/commandrunner -run '^TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets$' -count=1`.
- Repository gates: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`.

The timeout-prone monolithic `go test ./...` and `make verify` were intentionally left to CI per
the repository verification contract. No production connector declaration, provider call, or
provenance evidence was changed.
