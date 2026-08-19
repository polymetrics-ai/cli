Refs #4015

## Intent

Certify real declaration-owned connector commands with inspectable, bounded evidence rather than inferred reachability.

## What Changed

- Added a connector-parameterized live-certification runner with no connector-specific branches. It derives candidates, current live eligibility, command arguments, request shapes, and declared non-secret credential defaults from each bundle.
- Persisted 38 validated schema-v2 GitHub `observed_operations` evidence records immediately after the corresponding real command passed its declared produced-value assertions.
- Added sanitized per-operation receipts for all 122 executed candidates and the unchanged second-connector proof.

## Live Results

- GitHub eligible sweep: `executed=122`, `certified=38`, `provider_refused=80`, `missing_fixture=4`, `product_defect=0`.
- Every accepted record passed `go run ./cmd/connectorgen certification-matrix --check` before the next command started; the final matrix check also passed.
- Freshchat used the same script unchanged and recorded its definition-owned no-candidate non-pass: `executed=0`, `missing_fixture=1`.
- The definition-only runner path passed across all 36 command-surface connectors.

## Verification

- `node --check scripts/certify-connector-live.mjs`
- `go test -timeout 20m ./cmd/connectorgen`
- `go test -timeout 20m ./internal/connectors/commandrunner`
- `go test -timeout 20m ./internal/cli`
- Repository gates: `make fmt tidy-check vet build docs-check-no-build smoke-no-build lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check`

## GSD / TDD

The required lifecycle was executed inline because this direct-PR brief forbids compatible isolated roles. Planning, red/green ledger, run state, and verification evidence are in `.planning/phases/issue-4015-generic-live-certification-runner-r1/`.

Loaded required skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## Safety

The disposable token was inherited from an environment variable set by command substitution; it is absent from argv, terminal output, receipts, and evidence. Evidence fingerprints command/request/response scalar values with a repository-salted HMAC. Each runner-owned credential and local project was removed and cleanup was verified.

## Follow-up

After opening this PR, the recorded 403/404 permission refusals will be selected from the receipt for an App-credential retry, excluding all enterprise-scoped commands as directed.
