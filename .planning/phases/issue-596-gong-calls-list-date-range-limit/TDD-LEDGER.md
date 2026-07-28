# TDD Ledger — issue 596 Gong calls list correction

## Resume checkpoint

- [x] Reread ship instructions and captain decision after compaction.
- [x] Revalidated GSD adapter with `scripts/gsd doctor`; `programming-loop` fallback remains recorded in `PLAN.md`.
- [x] Verified current authoritative Gong OpenAPI docs before implementation finalization; private comparative notes remain outside tracked/public surfaces.
- [ ] Next active slice: write fail-first request-shape, validation, limit/cursor, and help assertions before implementation.

## Red

- [x] Add fail-first tests for `calls list --from` / `--to` request shape (`TestGongCallsListDateFlagsMapToQuery`; failed with unknown flags).
- [x] Add fail-first tests for invalid timestamp and invalid range rejection before HTTP (`TestGongCallsListRejectsInvalidDateFlagsBeforeHTTP`; failed with unknown flags).
- [x] Add fail-first tests for `--config start_date` compatibility and explicit `--from` precedence (`TestGongCallsListFromFlagOverridesStartDateConfig` and engine precedence test; failed with unknown flag / config overwrite).
- [x] Add fail-first tests for `--limit` counts across cursor pages: 1, below boundary, boundary, above boundary (`TestGongCallsListLimitCapsEmittedRecordsAcrossCursorPages`; added with request-count assertions).
- [x] Add fail-first/help assertions for `pm gong calls list --help` command-specific flags and output-cap wording (`TestGongCallsListHelpDocumentsDateFlagsAndLimitOutputCap`; failed with no command-specific flags / old limit wording).

## Green

- [x] Implement command flags, validation, query precedence, and docs.
- [x] Review fix: reject present-but-empty Gong calls list date flags before HTTP.
- [x] Review fix: validate `--to` against effective `start_date` fallback when `--from` is absent.
- [x] Targeted tests pass:
  - `go test ./internal/connectors/commandrunner -run 'Test.*(Gong|OperationDirectRead|DirectRead)' -count=1`
  - `go test ./internal/connectors/engine -run 'TestRead(StartConfigIncrementalRequestParamDoesNotOverrideExplicitRequestQuery|IncrementalRequestParamOverridesRequestQueryCollision|IncrementalLowerBoundFallsBackToStartConfigKey)' -count=1`
  - `go test ./internal/cli -run 'Test(Gong|DynamicConnectorHelpAndBareNamespace|GongCallsListHelpDocumentsDateFlagsAndLimitOutputCap)' -count=1`
  - `go test ./cmd/connectorgen -run 'TestGong' -count=1`
  - `go test ./internal/connectors/conformance -run 'TestConformance/gong' -count=1`
  - `go run ./cmd/connectorgen validate internal/connectors/defs`

## Refactor

- [ ] Keep scope narrow; no participant stream, live checks, or unrelated Gong endpoint edits.
- [ ] Re-run formatting and validation after docs generation.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `no-mistakes`.
