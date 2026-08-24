## Refs

Refs #4329

## Intent

Permit an explicitly declared, source-cited non-mutating operation to remain
intentionally read-only without being misreported as a missing foundation.
Mutation operations remain outside this feature and are rejected when marked
`read_only`.

## What Changed

- Added the closed `api_surface.operation.model: read_only` member.
- Bound its declaration to a matching non-mutating source method/path plus a
  fixed named policy and reason.
- Rejected mutating or executable read-only declarations.
- Projected read-only evidence separately, with connector/policy rollups.
- Recorded red/green and manual-inline GSD lifecycle evidence under
  `.planning/phases/issue-4329-read-only-source-coverage-r1/`.

## Testing

- Focused source-projection, operation-evidence, and engine model tests.
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen`
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/engine`
- `GOFLAGS=-p=3 go run ./cmd/connectorgen operation-evidence --check`
- `GOFLAGS=-p=3 go run ./cmd/connectorgen validate internal/connectors/defs/sentry`
- `GOFLAGS=-p=3 go run ./cmd/connectorgen validate internal/connectors/defs/vercel`
- `GOFLAGS=-p=3 make verify`

Frozen GitHub source lock measured `3,420,025` bytes / `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`; descriptor measured `43,354,021` bytes / `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.

## Delivery and review

- GSD lifecycle ran inline because compatible isolated GSD workers were not
  available; the resolved prompts and red/green evidence are recorded in the
  phase files.
- Required Go skills loaded: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-naming`, `golang-code-style`, `golang-lint`, and
  `golang-documentation`.
- CLI help/manual/website parity is not applicable: this changes only the
  developer-only `connectorgen` declaration/evidence contract, not `pm` user
  command behavior.
- Claude automatic review is pending on PR creation; no Copilot fallback is
  requested.

## Safety and follow-up

No credentials, provider calls, or writes are added. Sentry and Vercel
source-cited POST/PATCH operations without actions remain visible inputs for
`cli-mutation-disposition-foundation-r1`; no connector declaration suppresses
them here.
