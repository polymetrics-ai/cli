# Issue #4291 — connector parity batches 6 and 7

## Task Delivery Header

- Issue: Closes #4291 — chore(connectors): map parity batches 6 and 7
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` with the declared source locks, six-class disposition ledgers, local validation, and repository connector gates green.
- Working branch: fm/cli-map-batch67-r1
- Task: Add a credential-free, public-source-locked six-class parity map for `close-com`, `outreach`, `salesloft`, `copper`, `zoho-bigin`, `klaviyo`, `braze`, `customer-io`, `intercom`, `freshdesk`, `segment`, `activecampaign`, `iterable`, `help-scout`, `gorgias`, `service-now`, `chatwoot`, `chargebee`, `square`, and `braintree`. Every authoritative provider operation receives exactly one disposition row. `enabled` requires an actual API-surface command or typed write-action binding: an unbound ETL stream remains `declaration-pending`. Typed write action endpoints remain enabled `direct_write`; their reverse-ETL eligibility is a separate attribute blocked by `generic-typed-destination-executor`.
- Verification: Run the ledger-invariant checker, `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`, the relevant package tests, and `make verify` gates individually, including `connector-boundary` through a bounded poll.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Each in-scope connector has a pinned, public-source lock | live | The source-lock file names every `api_surface.json` endpoint and records public document URL/hash/bytes; removing a lock makes the checker fail. |
| Every documented operation has one six-class disposition | live | The checker compares endpoint identities to ledger rows, rejects duplicates/missing rows, and counts every documented DELETE. |
| Reasons describe present engine capability accurately | live | The checker rejects a typed write action classified as `reverse_etl`, an enabled operation without a command/action binding, and verifies the separate reverse-ETL foundation-gap attribute against the supplied destination-executor evidence/minimal change. |
| Generator and connector surfaces remain valid | live | `connectorgen validate` and `surface-sync --check` operate on the real changed definition directories. |

## Lifecycle record

- Required GSD commands resolved with `scripts/gsd sources` and prompts generated for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- Inline/manual fallback: `gsd-sdk query init.phase-op connector-parity-batches-6-7-r1` reports `phase_found: false` because `.planning/ROADMAP.md` intentionally delegates connector work to the issue-first canon. The canonical delivery contract also forbids spawning GSD roles. This issue-local phase record executes the same discuss → TDD plan → execute → verify → review sequence inline.
- Discuss decisions locked by issue #4291 and the repaired brief: no credentials or provider API calls; source material is public documentation only; an authoritative provider specification or complete rendered reference is the denominator; no legacy `api_surface.json` entry bounds a recovery remap; un-authored operations are `declaration-pending`; deletes are not unsafe solely because they delete; ETL is source-declaration-pending when absent; typed write actions are enabled `direct_write`; their reverse-ETL eligibility uses the `generic-typed-destination-executor` gap stated in the 2026-08-19 correction.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## TDD slices

1. **RED — batch 6 source-lock and ledger invariants:** prove the ten locks/ledgers are absent and the invariant checker cannot pass.
2. **GREEN — batch 6 map:** materialize the ten connector-local locks and ledgers, then prove exact endpoint coverage, class assignment, delete accounting, and honest ETL/reverse-ETL dispositions.
3. **RED — batch 7 source-lock and ledger invariants:** prove the ten locks/ledgers are absent and the invariant checker cannot pass.
4. **GREEN — batch 7 map:** materialize the remaining ten maps and repeat the real invariant check.
5. **REFACTOR / review:** inspect generated JSON for source provenance and reason vocabulary, then run the repository gates without widening scope beyond these definitions and the issue evidence.
6. **SOURCE-LOCK RECOVERY — captain defect 2026-08-19:** hold PR #4296. Audit all 20 owned connectors against each provider's complete machine-readable specification, complete rendered reference, or explicit dynamic-instance basis. Replace every incomplete public-documentation pin; record `counts.total` plus per-method counts, replace self-referential `declared_percent` with `operations_found` and `coverage_confidence`/basis, and regenerate the API surface plus every documented-operation ledger row from the corrected denominator before requesting any PR progress. Record every connector's old/new count and basis, including a verified no-change result.

## Commit checkpoints

- Planning evidence and documented RED baseline.
- Batch 6 source locks and ledgers after its green validation.
- Batch 7 source locks and ledgers after its green validation.
- Review-fix checkpoint only if the required review finds an actionable defect.
