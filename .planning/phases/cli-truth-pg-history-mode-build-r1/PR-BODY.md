## Intent

Direct Firstmate delivery for audit finding F2: make the existing PostgreSQL
`incremental_dedupe_history` claim true. No GitHub issue number was supplied
with this direct-PR task. Base: `integration/4015-mvp-flat-r1`.

Refs #4015

Exactly one connector is implemented: `postgres`. No shared runtime, schema,
or unrelated connector implementation was changed.

## Rebase resolution — #4186

Rebased onto the current `integration/4015-mvp-flat-r1` head
`281560ca14f80df7e2c473edb350d133d0af8b98` after #4186 merged. The three
overlapping files were resolved explicitly:

- Regenerated `docs/connectors/catalog/all-connectors.json` rather than
  hand-merging; an immediate second generation was byte-stable.
- Kept #4186's definition-owned `source_bindings` coverage and its
  PostgreSQL-to-GitHub route cases in `internal/app/transport_composition_test.go`,
  alongside the PostgreSQL history executor-resolution case.
- Kept both #4186 PostgreSQL-to-GitHub binary tests (deterministic and opt-in
  live) in `internal/cli/postgres_transport_binary_integration_test.go`, and
  added the independent PostgreSQL history update/replay binary proof next to
  them.

The full local gate set below was rerun after the rebase; all checks passed on
the rebased tip.

## What Changed

- Declared history mode and its managed `dedupe_history` action on PostgreSQL's
  outer source and destination transport contracts, and regenerated the
  connector catalog projection.
- Sealed the native PostgreSQL history route from both legs' typed database
  declarations before adapter I/O; retained history primary keys and the
  durable source page position for the existing conditional order fence.
- Added registry admission, native route/key/position, and fresh-built-binary
  PostgreSQL source-to-separate-target tests.

History means an observed mapped source row plus `_valid_from`, `_valid_to`,
and `_is_current`. The declared primary key identifies a lineage; a newer
source cursor closes its current version and inserts the next current version
in one PostgreSQL transaction. Equal/late or identical replay is a no-op;
omission is not deletion. This preserves PR #4184's publish/receipt before
checkpoint behavior—no per-page checkpoint was introduced.

## TDD and GSD Evidence

- **Red:** outer preflight refused `incremental_dedupe_history`; after outer
  admission, the real binary exposed missing history-route then history-key
  forwarding before target writes.
- **Green:** typed PostgreSQL route, declared key, and source-owned candidate
  position are forwarded; preflight resolves registered executors and the
  live binary proves initial/update/replay target state.
- Lifecycle: `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` were resolved and executed
  inline. The direct-PR runtime has no compatible Pi role runtime and repo
  policy forbids role spawning.
- Skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  and `golang-database`.

## Testing

Happy case: a fresh shipped `pm` creates a saved PostgreSQL history connection
and completes an approved live run; an independent query of the target sees
the initial current row and receipt.

Update case: a real source update at a larger cursor yields exactly two target
versions: v1 closed at v2's `_valid_from`, v2 current.

Edge/replay: a third approved run reads/loads zero rows and leaves the
independently queried target history unchanged.

Bad case: non-PostgreSQL history routes and a source lacking a typed definition
are refused before session/ledger mutation.

Passed locally:

- `go test -timeout 20m -count=1 ./internal/app` (230.728s)
- `go test -timeout 20m -count=1 ./internal/connectors/database`
- `go test -timeout 20m -count=1 ./internal/connectors/engine`
- `go test -timeout 20m -count=1 ./internal/connectors/native/postgres`
- `go test -timeout 20m -count=1 ./internal/synctransport`
- `go test -timeout 20m -count=1 ./internal/cli` (472.760s)
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -count=1 -tags=databaseintegration ./internal/cli -run '^TestPMBinaryExecutesPostgresIncrementalDedupeHistory$' -v` (30.70s)
- `go vet ./...`; `go build ./cmd/pm`; `connectorgen` validate, surface-sync,
  certification-matrix, and boundary; generated website docs twice byte-stable
- `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`, and
  `make release-workflow-check`; `scripts/verify-gsd-workflow`

Rebased-tip result: all commands above were rerun after rebasing. The fresh
live PostgreSQL history proof passed in 32.052s; catalog and website
generation were each repeated and byte-stable.

## CLI, Docs, and Safety

No CLI syntax, help text, or website prose changed: the published capability
was already accurately worded. `./pm docs generate --dir docs/cli` refreshed
the required generated connector catalog; website generation was byte-stable.

The live test uses the opted-in local container endpoint supplied by the task.
It does not print, store, or alter credentials. No new dependencies,
destructive external action, generic write capability, or reverse-ETL path was
introduced.

## Pipeline and Review

Rebased implementation commit: `d2bb755ce`; this delivery-record follow-up
documents conflict resolution and rerun verification. Branch:
`fm/cli-truth-pg-history-mode-build-r1`.

Manual standard review found no unresolved findings. Claude automatic review
is the primary GitHub route on PR open; no Copilot request was made. No work is
intentionally deferred.
