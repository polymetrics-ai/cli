# TDD Ledger — issue 596 Gong calls list correction

## Resume checkpoint

- [x] Reread ship instructions and captain decision after compaction.
- [x] Revalidated GSD adapter with `scripts/gsd doctor`; `programming-loop` fallback remains recorded in `PLAN.md`.
- [x] Verified current authoritative Gong OpenAPI docs before implementation finalization; private comparative notes remain outside tracked/public surfaces.
- [ ] Next active slice: write fail-first request-shape, validation, limit/cursor, and help assertions before implementation.

## CI follow-up — PR #597 linked issue guard

- [x] Revalidated GSD adapter with `scripts/gsd doctor`; `scripts/gsd prompt programming-loop init --phase pr-597-require-linked-issue --dry-run` remains unavailable (`unknown GSD command: programming-loop`), so manual-GSD fallback stays active.
- [x] Loaded required skills for the Go validation fix: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.
- [x] Red: added issueguard coverage for the PR body wording `Ship ... for issue 596` and kept ambiguous `Issue 123` text rejected; `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` failed on the new positive case before implementation.
- [x] Green: delivery issue wording now accepts `ship` phrasing and unprefixed `issue 596` numbers within the delivery phrase while preserving rejection of ambiguous or negated references.
- [x] Verification: targeted issueguard tests, direct `go run ./cmd/prissueguard` CI-shape invocation, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` passed.

## PR 597 architecture revision

- [x] Reconciled old no-mistakes run `01KYKYYAPSP34HVH9VH5ZKK3WN` through `no-mistakes axi status/help`; cancelled it before editing so no second validation run is active.
- [x] Red: add generic definition-validation tests for malformed flag format/empty/order/fallback declarations.
- [x] Red: add generic commandrunner tests for date-time format, present blank values, missing sides, config fallback, strict ordering, and definition-owned messages without provider-named runtime branches.
- [x] Green: remove Gong-specific shared-runner branch and implement generic CLI-surface validation declarations/interpreter.
- [x] Green: declare Gong calls-list constraints in `cli_surface.json` and preserve existing customer behavior.
- [x] Verify `git grep -n -E 'gong|Gong|fromDateTime|toDateTime|start_date' -- internal/connectors/commandrunner/runner.go` returns no matches.

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
