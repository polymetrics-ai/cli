# Plan — make PostgreSQL `incremental_dedupe_history` executable

## Task Delivery Header

- Issue: Audit finding F2 — PostgreSQL `incremental_dedupe_history` truthfulness repair (direct Firstmate task; no GitHub issue supplied).
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with the branch committed and local required verification recorded.
- Working branch: `fm/cli-truth-pg-history-mode-build-r1`.
- Task: Align PostgreSQL's declarative outer transport with its existing inner history implementation, prove registered preflight and a shipped-binary live PostgreSQL history update/replay, and preserve run-scoped publish-then-checkpoint atomicity.
- Verification: Targeted red/green tests; tagged live binary integration; `go vet`, `go build`, generated/snapshot checks, lint, connector boundary, website generation followed by a byte-stable repeat, GSD verification, and code review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Outer source and destination contracts admit history and resolve real registered executors | live | `app.Open` loads the shipped bundle and `Registry.Preflight` returns the native PostgreSQL source/destination references plus `dedupe_history`; without the descriptor repair it returns the prior source-mode refusal. |
| Saved history connection executes through the shipped binary against live PostgreSQL | live | A freshly built `pm` creates the connection and runs it after the normal plan/approval flow; independently querying the target finds the mapped row and durable receipt. |
| An update produces declared history | live | A second built-binary run after a real source update yields exactly two target versions for the key: first closed, second current, with touching validity timestamps. |
| Replay is safe | live | A later approved built-binary replay leaves the independently queried two-row target history unchanged; no row is duplicated or reopened. |
| Bad routes remain non-I/O refusals | fake | A database-driver fake is necessary because the assertion is precisely that unsupported source/destination combinations must never open a write session. It asserts the existing typed `DatabaseWriteHistoryRouteError` and zero session/ledger calls. |
| Advertised surfaces match the repaired capability | live | `connectorgen` validation/surface checks and two website generator runs leave no diff; the existing connector docs, metadata, certification matrix, and generated payload continue to name the same now-executable mode. |

Every live item asserts a state change or independently observed state, never merely a successful command.

## TDD slices

1. **Red — outer admission.** Change the production-composition test from the
   known history refusal to require real registered source/destination
   executors and the `dedupe_history` strategy. Run it against the unchanged
   descriptor and retain its failure.
2. **Green — declarative repair.** Add the mode to both outer lists and add
   `{"mode":"incremental_dedupe_history","strategy":"dedupe_history",
   "action":"managed_incremental_dedupe_history"}`. Update the exact
   intersection expectation. The existing factories and registry must then
   resolve without production Go changes.
3. **Red — binary route.** Add a tagged live PostgreSQL test that builds `pm`,
   saves a history connection, and attempts its first approved run. The
   descriptor repair exposed the remaining production seal: outer destination
   planning lacked a typed history route and history primary keys, causing the
   existing plan boundary to refuse before target writes.
4. **Green — end-to-end history.** Carry the immutable PostgreSQL database
   declaration through both the inner polling runner and outer destination;
   seal PostgreSQL→PostgreSQL before adapter I/O, retain history primary keys,
   and attach each durable page candidate tuple to its ordered records. Use the
   binary test to run a source row, update that key at a larger cursor, and
   replay. Separately query target PostgreSQL for history windows and row
   count.
5. **Regression.** Run existing history route rejection and direct adapter
   integration tests, then package/global gates. Confirm the run path retains
   PR #4184's one-receipt-before-checkpoint behavior.

## CLI/help/manual/website parity

- [x] No new CLI syntax or output format: `connections create` and `etl run`
  already accept the mode.
- [x] Bare namespace, help topic, command help, and invalid-action behavior:
  not changed; run relevant existing CLI tests rather than rewrite help.
- [x] Existing `docs.md`, `database.json`, certification matrix, and generated
  connector website data already state the desired behavior; no wording change
  is accurate or required.
- [x] `pnpm --dir website run gen:docs` ran twice; `git diff --exit-code --
  website` was clean after the second run.

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and
`golang-database`; the PostgreSQL runtime guidance and CLI parity reference
were read. GSD commands are executed inline because this direct-PR worker has
no compatible Pi role runtime and the repository forbids role spawning:
`discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` →
`code-review`.

## Verification commands

```sh
go test -timeout 20m -count=1 ./internal/app ./internal/synctransport ./internal/connectors/database ./internal/connectors/native/postgres
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -timeout 20m -count=1 ./internal/cli -run 'TestPMBinaryExecutesPostgres.*History|TestPMBinaryExecutesPostgresWarehousePostgres'
go vet ./internal/app ./internal/synctransport ./internal/connectors/database ./internal/connectors/native/postgres ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen validate
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-matrix --check
go run ./cmd/connectorgen boundary
pnpm --dir website run gen:docs
pnpm --dir website run gen:docs && git diff --exit-code -- website
make tidy-check
make lint
make docs-check-no-build
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```
