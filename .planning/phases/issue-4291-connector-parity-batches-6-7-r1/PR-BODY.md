Closes #4291

> **Held — source-lock completeness recovery in progress.** The original maps correctly classify their then-current `api_surface.json` rows, but that denominator was incomplete for several providers. Do not merge until the recovery checklist is complete.

## Summary

- Adds public-documentation source locks and declaration-disposition ledgers for batches 6 and 7 (20 canonical connector IDs); source-lock completeness recovery is now correcting the provider denominator before merge.
- Salesloft is the first comprehensive recovery slice: its former 12-operation documentation-index map is replaced by a 315-page provider-reference crawl with **211** cited operations (120 GET, 50 POST, 23 PUT, 18 DELETE), and both `api_surface.json` and its full disposition ledger were regenerated from that inventory.
- Makes enabled status reachable: an endpoint is enabled only with a bound API-surface direct-read command or typed write action. The former **2,099 documented / 575 enabled / 348 commands / 448 writes / 211 deletes** is explicitly a superseded legacy-crosswalk total, not a provider-surface claim.
- Corrects write semantics: typed actions are `direct_write`; reverse ETL is a separate, currently blocked attribute with `generic-typed-destination-executor` and no fabricated transport binding.
- Captures Close.com as `close-com` and ServiceNow as `service-now`; both public source retrievals succeeded.

## Delivery record

- Issue-first GSD sequence completed inline. The adapter has no roadmap phase for #4291 (`phase_found: false`) and the canonical contract forbids spawning GSD roles; the discuss, TDD plan, execute, verification, and manual-review artifacts are in `.planning/phases/issue-4291-connector-parity-batches-6-7-r1/`.
- TDD evidence: source-lock/ledger absence checks supplied the red state; the strict complete-map invariant check supplied green state. The recovery adds a 211/211/211 Salesloft source-lock/API-surface/ledger invariant and rejects `declared_percent`.
- Required skills used: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- Safety: all provenance came from public documentation. No credentials, provider write calls, engine changes, or command-surface changes were made.

## Verification

Passed: `connectorgen validate`, `surface-sync --check`, changed-package and `internal/cli` tests, `go vet ./...`, `go build ./cmd/pm`, docs validation, smoke/lint/agent-contract/connectorgen make gates, generated-ledger snapshot checks, certification checks, connector boundary, and release/canon scripts. Full commands and results: `.planning/phases/issue-4291-connector-parity-batches-6-7-r1/VERIFICATION.md`.

The repository instruction specifically asks per-command agents not to run aggregate `go test ./...` or `make verify`; their full-suite coverage is left to CI after the scoped tests and individual non-suite gates above.

## Review

Manual diff review found no actionable defects. Claude automatic PR review is expected on PR creation; no duplicate on-demand request is needed unless new commits require it.
