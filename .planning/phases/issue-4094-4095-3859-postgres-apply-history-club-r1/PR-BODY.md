Closes #4094
Closes #4095
Refs #3859

## Intent

Land the PostgreSQL apply/history club on `integration/4015-mvp-flat-r1` by
preserving the staged history target and CDC delete binding while closing the
audited #3859 residual: the registered database polling adapter could not plan
`incremental_dedupe_history` because it supplied no sealed history route.

## What Changed

- Obtain the immutable database definition from the exact polling source
  resolved by preflight, bind it to the descriptor source identity, and seal
  its driver with the destination definition only for history plans.
- Exercise v1, v2, restart replay, and a CDC-derived tombstone through
  `DatabasePollingApplyExecutor`, checking both real PostgreSQL row state and
  changed durable delivery IDs.
- Record the required GSD/TDD plan, red/green traces, acceptance table,
  verification, and deep inline review.

No file under `internal/synctransport`, descriptor validation, or the transport
registry changed; #4093 owns that work. #4125 was not touched. #4158 was a
pre-existing base symptom and was not taken as a separate fix; the current
tagged PostgreSQL package, including its named live control-assertion test,
passes with this adapter residual fixed.

## TDD and GSD

- `scripts/gsd prompt discuss-phase 4094 --auto`
- `scripts/gsd prompt plan-phase 4094 --tdd`
- `scripts/gsd prompt execute-phase 4094 --interactive`
- `scripts/gsd prompt verify-work 4094`
- `scripts/gsd prompt code-review 4094 --depth=deep`

The generated workflows were executed inline because the repository's
single-worker contract forbids spawned lifecycle roles for this combined
delivery.

Red: the live adapter returned a zero acknowledgement and the missing-route
plan refusal before its first history write.

Green: the same live test returned durable acknowledgements and queried exact
validity-window, replay, restart, receipt, and soft-delete-close state.

Refactor: route authority stays definition-owned on the resolved source; no
second caller-supplied source identity was added to target configuration.

Skills used: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`golang-database`, and the five required GSD lifecycle skills.

## Testing

- PASS: focused default engine/database/PostgreSQL packages
- PASS: focused live history, live workset, CDC tombstone, and typed
  non-PostgreSQL route matrix
- PASS: race detector on all three affected packages
- PASS: full tagged `internal/connectors/native/postgres` package (47.189s)
- PASS: scoped `go vet`, `go build ./cmd/pm`, and all applicable individual
  repository gates (`tidy-check`, lint, docs/smoke, agent contract,
  connectorgen validation/surface sync, connector boundary, release workflow)
- PASS: `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1`

The requested tagged `./...` run had one parallel container teardown/LSN
inspection failure in the untouched PGOutput harness. The exact test passed on
isolated rerun (10.22s), then the full tagged PostgreSQL package passed. Per the
high-load instruction, the broad suite was not repeated.

Every acceptance criterion and its observable state assertion is recorded in
the phase `VERIFICATION.md`. CLI help/manual/docs/website parity is not
applicable because no public command surface changed.

## Delivery and review

- Base: `integration/4015-mvp-flat-r1`
- Head: `fm/cli-pg-apply-history-club-r1`
- Checkpoints: plan, red test, green implementation, verification/review
- Deep inline review: no findings
- Claude route: automatic review expected on PR open; pending
- Copilot route: not requested (fallback only)
- No credentials, generic SQL/API/shell write surface, dependency, destructive
  external action, or reverse-ETL approval bypass was introduced.
