# Plan — Google Search Console documented-operation parity resume

## Objective

Rehydrate PR #3555's Google Search Console connector work onto current `origin/main`, correct
all current validator findings, and make every operation in Google's published Search Console
reference genuinely reachable through the appropriate safe connector surface.

## Scope and ownership

- Owned definition scope: `internal/connectors/defs/google-search-console/**`.
- Connector-specific hook scope: `internal/connectors/hooks/google-search-console/**`.
- Generated manual/catalog/website artifacts are regenerated late, one connector at a time.
- Do not edit shared engine, validator, schema, generated-index, or other connector paths. PR #3555
  carried old shared changes; current-main versions are retained under the rehydrate-don't-resurrect
  contract.

## GSD and orchestration record

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase cli-google-search-console-parity-resume-r1 --dry-run`
  reported that `programming-loop` is not an exposed adapter command. The repo-local
  `scripts/programming-loop.mjs` helper is absent. This phase therefore uses the permitted
  manual-GSD fallback, following the universal runtime loop and this phase record.
- Execution decision: `local_critical_path`. This is one coupled connector bundle/hook repair
  lane, and the current task does not request delegated workers.
- Skills loaded: `gsd-programming-loop`, `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, and `golang-documentation`.

## Test-first slices

1. Capture the current-main validator failure against the rehydrated bundle. Record all nine
   reported findings (four stream target mismatches and five missing required flags) in
   `TDD-LEDGER.md`.
2. Build a provider-field research matrix from Google's Discovery document and published REST
   reference. Cite every request path/query/body field with source URL/section, evidence type,
   confidence, and requiredness rationale. This is research evidence only until the shared citation
   convention lands; do not invent a competing bundle format.
3. Correct stream/operation targets and CLI required-flag mappings in the Google Search Console
   bundle. Re-run the focused validator until it reports zero findings.
4. Establish the actual documented operation inventory from provider-owned published reference,
   distinguish source operations from convenience streams/commands, and leave no documented
   operation `planned`. Express record-shaped mutations through `writes.json`; do not mark
   `rest_write` operations implemented.
5. Check current `origin/main` late for the shared citation-convention change. If it has landed,
   encode the researched citations in its canonical shape; otherwise retain the required research
   matrix rather than inventing a competing bundle format. Regenerate connector artifacts and
   verify real command registration plus representative fixture-backed requests.

## Safety and reporting

- No credentials, live provider calls, reverse-ETL execution, raw HTTP tooling, new dependencies,
  shared-engine edits, or main-branch merge.
- Deletes remain typed, approval-gated, destructive, and idempotent only where provider behavior
  supports it.
- Final report states Google’s true documented operation total, reachable count, exact reasons for
  every non-reachable operation, field-citation coverage, stale blocker prose retired, and all
  executed gate output.

## Required verification

- `go run ./cmd/connectorgen surface-sync`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/connectorgen validate internal/connectors/defs/google-search-console`
- `go test ./internal/connectors/conformance/...`
- `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`
- `go test ./internal/cli/...`
- targeted `go vet` / `go build` for changed packages and `go build ./cmd/pm`
- `cd website && pnpm run gen:website-data`
- built `pm` help and representative Google Search Console command execution against local fixture
  infrastructure; no credentialed provider call.
