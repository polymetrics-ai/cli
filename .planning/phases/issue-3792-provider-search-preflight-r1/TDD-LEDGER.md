# TDD ledger — issue #3792 provider-search runtime preflight

| ID | Contract | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | An implemented operation-backed direct-read command cannot preflight through a generic reader alone. | `go test ./internal/connectors/commandrunner -run '^TestPreflightOperationDirectReadRejectsNonExecutableOperationMetadata$' -count=1` failed on the unchanged runner: `Preflight error = nil, want loaded operation rejection`. | Pending. | Must prove `OperationDirectRead` is not dispatched after rejected preflight. |
| R2 | A loaded operation must be executable as `rest_read` or `provider_search`, have a matching fixed method/path/API-surface row, and positive cap. | Pending: engine table cases. | Pending. | Factor the executor and preflight through one admission helper. |
| R3 | Command operation, method, path, and policy must exactly match loaded metadata. | Pending: method/path/policy table cases. | Pending. | No `connectorgen` duplicate; real runtime sweep exercises this path. |
| R4 | Provider-search input stays closed/bounded and invalid input never issues a request. | Existing `provider_search_test.go:248-314` is retained; this phase adds no new executor. | Existing coverage remains green. | Verify focused engine suite. |
| R5 | Full-content output policy is not invented by this lane. | Changed-path inspection must show no #3771/#3852-owned code or policy/schema change. | Pending. | Record deferred ownership in final verification. |

## Red-test rule

Each new behavior test runs against unchanged production code first. The observed failure is retained
as evidence; no skip, deletion, or weakened assertion may make the claim true.
