## Intent

Refs #4069 and Refs #3897.

Deliver the existing #4069 warehouse collision corrections on the current
stacked draft #4071 without changing its base, state, review readiness, or
merge status. Preserve generic SQL and the typed same-connection local
collision boundary. Do not add credentials, connector-specific shared-code
branches, transport registration, or provider work.

## What Changed

- Restored destination-scoped admission for the legacy same-owner
  case-equivalent collision preflight in `App.RunETL`.
- Restored `TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL`, which
  proves an unrelated non-local ETL is unaffected by a legacy local collision.
- Preserved the existing local true-positive test and its typed
  `*warehouse.SameOwnerCaseEquivalentTableError` pre-mutation refusal.

## Risk Assessment

Low. The condition reuses the existing credential-free
`LocalWarehouseMaterializer` abstraction: local destinations retain the guard;
non-local destinations bypass only an irrelevant local-inventory check. No
connector name or warehouse literal is introduced in shared code.

## Testing

Target binary: `5a4f3d23645d3046d02055273fedd9a8b9dd67d1`

Red then green:

```text
go test -timeout 20m ./internal/app -run '^TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL$' -count=1 -v
# RED on source 3b75f4a62fd8d743ec883a5b824164374f661857:
# RunETL(non-local) error = connection "acme" configures case-equivalent local warehouse destination tables ...
# exit 1

go test -timeout 20m ./internal/app -run '^(TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL|TestLegacySameOwnerCaseEquivalentInventorySurvivesOpenAndFailsBeforeMutation)$' -count=1 -v
# PASS: non-local regression and local typed pre-mutation guard

go test -timeout 20m ./internal/app -count=1
# PASS: polymetrics.ai/internal/app (98.497s)
```

Additional local gates passed: `go vet ./...`, `golangci-lint run
./internal/app` (0 issues), `go build ./cmd/pm`, `make tidy-check`, `make
lint`, `scripts/tests/release-target-parity.sh`, `go run
./cmd/agentcontractgen check`, `go run ./cmd/connectorgen surface-sync
--check`, and `scripts/verify-gsd-workflow`.

## GSD/TDD Evidence

The inline/manual single-worker fallback resolved and executed the required
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts. Required skills loaded: `golang-how-to`,
`golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and
`golang-code-style`. Exact RED/GREEN and verification records are under
`.planning/phases/issue-4069-flow-case-equivalent-unique-tables-r1/`.

## Delivery Boundary

This PR remains draft and stacked on `feat/3897-flow-connection-scope-nm`.
It is not retargeted, merged, closed, or marked ready for review. A fresh
no-mistakes run and natural exact-head checks remain the next delivery gate.
