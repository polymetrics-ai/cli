# Code review — Issue #3973 mapping contract completion

## Scope reviewed

- `mapping_contract.go`: V1 mapping construction, defensive copies, type-plan
  verification, value projection, and inverse round-trip checks.
- `database_write_input.go`: explicit row-delete envelope/input validation and
  bounded batch construction.
- `database_write_session.go`: sealed plan equality, compatibility bridge,
  pinned-session dispatch, and named receipt type.
- Focused tests and GSD evidence.

## Findings

No critical, warning, or informational findings remain.

The review specifically confirmed that no new SQL, DDL, connection, credential,
arbitrary relation, generic write tool, capability flip, or physical-absence
delete path was added. Validation rejects non-row-delete tombstones before
session opening; plan/input count mismatches are rejected before approval
consumption; and errors do not render records, tombstone keys, or other
provider values.

## Evidence

- `go test -timeout 20m ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/... -count=1`
- Focused race test for mapping and tombstone paths
- Targeted `go vet` and all scoped `make verify` component gates
- `scripts/gsd prompt code-review issue-3973-mapping-contract-r2` generated
  and executed inline under the no-delegation fallback.

Automated GitHub review is pending the direct sub-PR opening; trusted-author
Claude automatic review is the selected primary route.
