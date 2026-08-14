## Intent

Refs #4069 and Refs #3897.

Deliver the existing #4069 warehouse collision corrections as a child PR from
`fm/cli-4071-collision-scope-fix-r1` onto
`fix/4069-flow-case-equivalent-unique-tables-r1`. Preserve generic SQL and the
typed same-connection local collision boundary. Do not change #4071 or add
credentials, connector-specific shared-code branches, transport registration,
or provider work.

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

Verified production commit: `a49ae952d52a6edf9786cc90e02825e2414f5b44`

Child PR actual head: pending publication by the delivery owner; at delivery
end this body and `RUN-STATE.json` must name that PR's actual head SHA.

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

This child PR is not yet published. Its delivery owner must create it with base
`fix/4069-flow-case-equivalent-unique-tables-r1` and record its actual head
SHA at delivery end. PR #4071 must not be retargeted, merged, closed, body
patched, or marked ready for review.
