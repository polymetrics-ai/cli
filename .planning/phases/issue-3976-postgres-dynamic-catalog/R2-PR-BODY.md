Refs #3976

## Intent

Make PostgreSQL's shipping source transport use the shared resumable polling
executor, rather than its private bounded-snapshot page loop.

## What changed

- The PostgreSQL source declaration/factory now selects
  `native_database/postgres_polling_watermark` for all supported non-CDC
  source modes. The former production private loop and factory are removed.
- The App → transport request carries the configured stream cursor. PostgreSQL
  requires a per-stream cursor plus one distinct non-null unique tie-breaker,
  validates them against the live catalog, and rejects missing, unknown,
  nullable, or unsupported cursor input before a page query.
- PostgreSQL builds an effective catalog-bound polling declaration and invokes
  the shared preflight/source executor. Checkpoints resume from their committed
  tuple; invalid protocol/barrier checkpoints and stale catalog fingerprints
  return typed rebootstrap reasons rather than restarting at page one.
- A narrow shared-executor correction commits a valid zero-row observation
  without creating or advancing a tuple, allowing a resumed empty poll to
  complete durably.
- Generated connector manuals/catalog and website data were regenerated. The
  static polling manifest remains intentionally `planned` because it cannot
  name a table-specific cursor/type/tie-breaker; the runtime source declaration
  is implemented and catalog-bound.

## TDD evidence

- Happy path: `TestPostgresPollingTransportResumesFixtureCursor` and
  `TestPMBinaryExecutesPostgresFixturePollingResume` assert exact records and
  a resumed zero-row result through the compiled `pm` binary.
- Bad path: `TestPostgresPollingTransportRefusesMissingPerStreamCursorBeforeIO`
  and `TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead`
  assert named refusals before source pages/checkpoint mutation.
- Edge cases: `TestPostgresPollingReadPlanRefusesNullableCursor`,
  `TestPostgresPollingTransportRefusesInvalidCheckpointWithoutRestart`, and
  `TestPostgresPollingTransportRefusesStaleSchemaCheckpointBeforePageRead`.
- Shared zero-row behavior was red-tested first by
  `TestPollingSourceExecutorCommitsEmptyPageWithoutInventingCursor`.

## Verification

- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/native/postgres`
- `go test -race -timeout 20m ./internal/connectors/native/postgres`
- Focused App, synctransport, CLI binary, and inspection tests
- `go vet` on changed packages; `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check`
- `go test -tags=databaseintegration -run '^$'` for PostgreSQL and CLI
  compiled the opt-in live path without starting it.

## Delivery and safety

- GSD lifecycle: `discuss-phase → plan-phase --tdd → execute-phase →
  verify-work → code-review` was executed inline because #3976 is a historical
  non-roadmap phase and the canonical contract prohibits role delegation. The
  plan, ledger, verification, and review records are under
  `.planning/phases/issue-3976-postgres-dynamic-catalog/`.
- Skills: `golang-how-to`, `golang-cli`, `golang-database`,
  `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-context`, `golang-concurrency`, and `golang-lint`.
- No secrets are read or emitted; no generic SQL/write interface is added.
- Live PostgreSQL proof is **pending**. The shared container runtime was not
  started or restarted because this task explicitly forbids that recovery
  action when the runtime is unreliable.
- Automated-review route: `claude_auto` pending on this non-default-base PR;
  parent/human coverage is the fallback. Copilot was not requested.
