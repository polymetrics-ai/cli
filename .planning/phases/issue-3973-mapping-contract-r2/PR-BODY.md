Closes #3973

## Summary

- seals a typed, ordered `MappingContractV1` into every database write plan;
- adds bounded `DatabaseWriteInput`/`TombstoneEnvelope` so only explicit row-delete tombstones reach a pinned session;
- returns `DeliveryReceiptV1` from sessions while retaining ledger persistence as the separate acknowledgement gate.

## Red / green evidence

- Red: `traces/red-run.txt` proves the missing V1 mapping API; `traces/tombstone-red.txt` proves the base had no plan/input/session tombstone path.
- Green: lossless `int32 -> int64 -> int32` round trip, narrowing/unrepresentable refusal, mapping-bound approval refusal, explicit tombstone deletion, and mismatched-tombstone pre-session refusal.

## GSD and skills

- Inline/manual lifecycle: `scripts/gsd prompt discuss-phase 3973`, `plan-phase issue-3973-mapping-contract-r2 --tdd`, `execute-phase issue-3973-mapping-contract-r2`, `verify-work issue-3973-mapping-contract-r2`, and `code-review issue-3973-mapping-contract-r2`.
- Inline fallback is required because this issue is outside numbered roadmap phases and the canonical worker forbids role spawning.
- Skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-database`.

## Verification

- `go test -timeout 20m ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/... -count=1`
- `go test -race -timeout 20m ./internal/connectors/database -run 'Test(MappingContractV1|DatabaseWritePlanSealsMapping|DatabaseWriteExecutor.*Tombstone)' -count=1`
- `go vet ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`

## Safety and deferrals

No credentials, live connection, DDL, SQL, generic write surface, capability promotion, CLI/documentation surface, physical-absence delete, or source checkpoint change is included. PostgreSQL-specific execution and live proof remain with #3982; immutable workset derivation and delivery remain with #3980 and #3983.

## Review coverage

Primary route: `claude_auto` on this trusted-author non-draft PR after opening. Fallback: none. Local GSD code review found no actionable findings; any GitHub review feedback will be dispositioned before integration.
