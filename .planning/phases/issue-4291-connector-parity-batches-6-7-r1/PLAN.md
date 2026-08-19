# Issue #4291 — connector parity batches 6 and 7

## Task Delivery Header

- Issue: Refs #4291 — chore(connectors): map parity batches 6 and 7
- Base branch: fm/cli-reverse-etl-destination-r1
- Merges into: fm/cli-reverse-etl-destination-r1 → main
- Delivery: PR #4296 remains open against `fm/cli-reverse-etl-destination-r1`; its twenty connector definitions, generated CLI/manual/website data, and required local gates are green.
- Working branch: fm/cli-map-batch67-r1
- Task: Reconcile the credential-free, public-source-locked inventory for `close-com`, `outreach`, `salesloft`, `copper`, `zoho-bigin`, `klaviyo`, `braze`, `customer-io`, `intercom`, `freshdesk`, `segment`, `activecampaign`, `iterable`, `help-scout`, `gorgias`, `service-now`, `chatwoot`, `chargebee`, `square`, and `braintree` against all seven delivery surfaces. Every documented operation must be faithfully modeled and user-reachable through the installed CLI when the existing closed contract supports it. Destructive, privileged, uncommon, binary, and non-certified operations remain reachable with their established safety/approval metadata; only a named technical contract gap can make an operation unsupported.
- Verification: Run the readiness-baseline invariant, `go run ./cmd/connectorgen validate`, `go run ./cmd/connectorgen surface-sync --check`, generated docs/website data checks, focused connector and CLI tests, `connector-boundary` through a bounded poll, and the non-suite `make verify` gates individually.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Each in-scope connector has a pinned, public-source lock | live | The source-lock file names every `api_surface.json` endpoint and records public document URL/hash/bytes; removing a lock makes the checker fail. |
| Every documented operation has one six-class disposition | live | The checker compares endpoint identities to ledger rows, rejects duplicates/missing rows, and counts every documented DELETE. |
| Reasons describe present engine capability accurately | live | The checker rejects a typed write action classified as `reverse_etl`, an enabled operation without a command/action binding, and verifies the separate reverse-ETL foundation-gap attribute against the supplied destination-executor evidence/minimal change. |
| Generator and connector surfaces remain valid | live | `connectorgen validate` and `surface-sync --check` operate on the real changed definition directories. |
| Every documented provider operation remains user-reachable unless a precise technical contract is absent | live | The seven-surface readiness ledger compares every source-locked direct-read/write/binary row against an exact command or typed action and rejects safety or certification as an unreachability reason. |
| Reverse-ETL declarations use the merged typed-destination contract | live | Bundle validation and runtime preflight load each connector-owned `sync_transport.json`, select only a named `writes.json` action, and reject an absent source binding, acknowledgement, or per-mode strategy before I/O. |

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
7. **REVERSE-ETL ACTION PREPARATION — captain freeze 2026-08-19:** while #4303 supplies no connector-neutral typed destination, identify every documented direct-write operation, retain / author only source-backed closed typed actions that can be candidates for its adapter, and record its readiness. Do not add a `transport_binding`, `sync_transport.json`, source binding, acknowledgement, or apply strategy until #4303's neutral factory and declaration schema land; all direct-write rows retain `generic-typed-destination-executor` as a reverse-ETL attribute.

## Commit checkpoints

- Planning evidence and documented RED baseline.
- Batch 6 source locks and ledgers after its green validation.
- Batch 7 source locks and ledgers after its green validation.
- Review-fix checkpoint only if the required review finds an actionable defect.

## Relaunch reconciliation — 2026-08-20

The preceding map is retained as provenance, but its prior generic destination gap and `main`
base are superseded. PR #4304 was merged locally as commit `d27d4bb64` and PR #4296 was retargeted
to `fm/cli-reverse-etl-destination-r1`; its declarative typed-destination factory is now the
available foundation. `READINESS-BASELINE.{json,md}` is the before-state for this reconciliation.

1. **RED — seven-surface baseline:** confirm the source-locked denominator has no source or
   destination transport claims, 1,465 direct-read rows have only 122 current exact command
   bindings, and 2,189 direct-write rows have only 444 exact typed-action bindings. Confirm the
   two current binary commands are missing binary ledger classifications. This is a reachability
   deficit, never a safety or certification exclusion.
2. **GREEN — connector-owned reconciliation:** add only provider-evidenced operation contracts,
   typed actions, fixtures, CLI bindings, source transports, and typed destination declarations;
   generate command/manual/website projections and prove preflight without credentials.
3. **REFACTOR / review:** remove stale `generic-typed-destination-executor` statements, make
   transport evidence connector-owned, classify binary operations separately, and preserve
   `provider_live_certification: pending` rather than treating fixture proof as certification.

The canonical contract forbids role spawning. The GSD lifecycle therefore runs inline with this
issue-local phase evidence; generated prompts were refreshed on 2026-08-20 after `scripts/gsd
doctor` and all five command sources passed `agentcontractgen check`.
