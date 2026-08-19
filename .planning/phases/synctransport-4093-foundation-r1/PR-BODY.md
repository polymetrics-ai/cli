Refs #4093

## Intent

Make production sync-transport composition definition-owned and connector-neutral,
without widening the closed issue-label action into a generic writer.

## What Changed

- Collect source and destination conformance evidence from each declaring
  `sync_transport.json` definition; a shared exact executor is registered once
  and dispatches using the preflighted request connector.
- Retire the fixed GitHub evidence constants and first-connector-bound
  production adapters. The declarative source is request-scoped and the
  issue-label destination remains a closed typed named-action adapter with
  planning, approval, acknowledgement, and read-back.
- Reject `change_capture` destination declarations. Capture remains on the
  source-only connection-warehouse path.
- Add the mechanical authoring guide at `docs/sync-transport-definition.md`.

## Evidence

- Red: new evidence-selection and factory tests first failed because the
  factory admitted one fixed evidence and lacked descriptor-derived factories;
  the shared source then failed because it was registered twice. The new
  destination-role test first failed because `change_capture` was accepted.
- Green: synthetic definition-owned registration runs through the generic
  App/orchestrator path without connector dispatch edits; a second declarative
  source declaration selects distinct evidence and preflights through the one
  production adapter. Unknown/malformed declarations, invalid CDC destination
  roles, altered evidence, and unbound sources refuse before I/O.
- Existing reconciliation and owned-stage cleanup tests prove unknown commit
  recovery does not replay a durable effect and cleanup is connection-owned
  and bounded.

## GSD and skills

Ran `scripts/gsd doctor`, resolved and generated prompts for `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`, and
recorded the inline/manual fallback because this runner cannot host compatible
isolated GSD workers. Required skills used: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`golang-documentation`, and `golang-lint`.

## Verification

Passed locally:

- `go test -count=1 -timeout 20m ./internal/app`
- `go test -count=1 -timeout 20m ./internal/synctransport`
- `go test -count=1 -timeout 20m ./internal/connectors`
- `go test -count=1 -timeout 20m ./internal/connectors/engine`
- `go test -count=1 -timeout 20m ./internal/cli`
- `go test -count=1 -timeout 20m ./cmd/connectorgen`
- `go test -race -count=1 -timeout 20m ./internal/synctransport ./internal/app`
- `go vet ./...` and `go build ./cmd/pm`
- `go run ./cmd/connectorgen validate`, `surface-sync --check`, and `boundary . --json`
- `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check connector-canon-check connectorgen-certification-matrix github-parity-artifacts-check`

The runner guidance defers a monolithic `go test ./...` to CI; it can exceed the
per-command window. The complete command/evidence record is in
`.planning/phases/synctransport-4093-foundation-r1/VERIFICATION.md`.

No CLI command/help/manual/website surface changed. No credentials were read,
stored, or printed. No generic HTTP, SQL, shell, or arbitrary write surface
was added.

## Review

Inline code review is recorded in the phase artifact. Claude automatic review
is the selected route on PR open; no Copilot fallback was requested. No
intentional follow-up or human decision is outstanding.
