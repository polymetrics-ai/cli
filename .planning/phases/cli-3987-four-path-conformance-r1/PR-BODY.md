## Intent

Refs #3987. Add the missing deterministic conformance matrix for the four canonical GitHub/PostgreSQL warehouse-mediated directions without changing the independently owned live-route or final-certification lanes.

## What changed

- Bound generated `api_to_api`, `api_to_database`, `database_to_api`, and `database_to_database` IDs to distinct production registry descriptors and persisted dispatch selections.
- Made the shared orchestration proof observable: destination plan resolves before source read; the source stages under its connection owner; the destination receives only the sealed receipt plus reopened workset; read-back precedes checkpoint.
- Accounted for every current `synccontract.Mode`. `incremental_dedupe_history` passes because merged PRs #4187 and #4188 made it executable. This intentionally diverges from stale #3987 prose that asked for a refusal.
- Proved `change_capture` as PostgreSQL's implemented CDC source contract; PostgreSQL's normal destination mode set rejects it before executor I/O, and its non-pass result has the concrete refusal reason and is excluded from the six-mode pass roll-up.

## Proof classification

The new matrix is deterministic simulated CI evidence, not a new live certification claim. Existing fresh-binary/live route proofs remain authoritative for the external paths: #4185 API→API, earlier authenticated GitHub API→PostgreSQL including the 90k-commit regression, #4186 database→API, and #4184 database→database.

- Happy: every named direction resolves its actual source/destination executors; all six currently executable modes preflight with their declared apply strategy.
- Bad: destination-shaped PostgreSQL `change_capture` preflight returns the specific closed-mode refusal before executor I/O; a non-pass result cannot join the roll-up.
- Edge: the scratch change retained schema validity while replacing the API→API admitted stream. Only the named `api_to_api` test went red, then passed after the exact binding was restored.

## TDD / GSD

- Red: `go test -count=1 -timeout 20m ./internal/app -run '^TestWarehouseMediatedFourPathConformance$'` failed because the conformance-contract helper did not exist.
- Green/refactor: the focused app and transport matrix tests pass; the ledger records the schema-valid direction-specific failure demonstration and restoration.
- Lifecycle: inline/manual `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` due to the canonical no-role-spawn delivery contract.
- Skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-code-style`, and `golang-lint`.

## Verification

- `go test -count=1 -timeout 20m ./internal/app ./internal/synctransport`
- `go test -count=1 -timeout 20m ./internal/connectors ./internal/synccontract`
- `go test -count=1 -timeout 20m ./internal/cli`
- `go test -count=1 -timeout 20m ./cmd/connectorgen`
- `go vet ./internal/app ./internal/synctransport ./internal/connectors ./internal/synccontract`
- `go build ./cmd/pm`
- Individual `make verify` gates, including lint, connector runtime preflight, generated-artifact checks, connector boundary, docs validation, smoke, and release workflow checks.
- `pnpm --dir website run gen:docs` twice with a clean diff after each run.

The aggregate `go test -timeout 20m ./...` and `make verify` were not run as one command because repository policy identifies command-harness timeout ambiguity; changed packages, consumers, `internal/cli`, and every non-test verify gate were run individually.

## Safety and scope

No credentials were read or recorded; no live provider/database write was performed. No connector definition, PostgreSQL profile/adapter, GitHub direct-read surface, generic transport registration, CLI dispatch/help, final certification publication, or `allStagesPassed` behavior was changed.

## Review coverage

PR [#4195](https://github.com/polymetrics-ai/cli/pull/4195) is open against `integration/4015-mvp-flat-r1`; an API-backed `gh-axi` base/head query returned exactly that PR. The PR-open automatic Claude review route is pending. This stacked PR's coverage record and any dispositions will be updated after GitHub reports review state.
