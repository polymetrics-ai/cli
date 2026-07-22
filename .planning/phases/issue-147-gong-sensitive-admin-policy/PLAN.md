# Gong sensitive/admin reverse ETL

Parent: #133
Issue: #147
Branch: `feat/133-gong-cli-parity`

## GSD path

- Adapter health: `scripts/gsd doctor` passed in this session.
- Requested workflow: `scripts/gsd prompt programming-loop ...` remains unavailable (`unknown GSD command: programming-loop` from earlier run); manual-GSD fallback is active and recorded.
- Planning prompt available: `scripts/gsd prompt plan-phase 133 --skip-research`.
- Required skills loaded: gsd-core, golang-how-to, golang-cli, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-testing, golang-documentation.

## Objective

Model Gong JSON mutations as typed reverse-ETL write actions with schemas, flags, risk/approval, destructive confirmation, and safe preview support.

## Constraints

- No secrets requested, printed, stored, or summarized.
- No credentialed Gong checks.
- No generic raw HTTP write, arbitrary GraphQL mutation body, generic shell write, or direct-write escape hatch.
- Reverse ETL remains plan -> preview -> approval -> execute; destructive actions require typed confirmation.
- Sensitive/admin/destructive operations are implemented as typed surfaces where the engine supports the request shape; non-JSON body/query/binary gaps remain typed operation metadata with bounded policies and exact blocker notes.

## Implementation plan

1. Add fail-first tests that assert the Gong surface has executable command coverage and safety policy coverage.
2. Extend shared direct-read output policy only as needed for bounded JSON redaction.
3. Generate/update Gong definitions from the public OpenAPI operation inventory.
4. Refresh connector docs/manual notes and planning state.
5. Run targeted validation and Go tests before broader gates.

## Verification target

- `go test ./internal/connectors/engine -run DirectRead -count=1`
- `go test ./cmd/connectorgen -run Gong -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

