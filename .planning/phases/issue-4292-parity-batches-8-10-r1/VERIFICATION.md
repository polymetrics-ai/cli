# Issue #4292 — verification checklist

## Reconciliation status — 2026-08-20

## Captain operation-evidence pre-merge gate — 2026-08-20

- [ ] **Blocked common foundation:** every provider-defined operation needs a
  source URL/version/hash trace, canonical mapping, enabled runtime
  reachability, generated CLI command, generated website row, and executable
  fixture/conformance evidence. The gate applies separately to ETL, reverse
  ETL, direct read, direct write, binary download, and binary upload.
- [ ] `N/A` must cite provider evidence that the capability is absent. Scope,
  tier, destructive, and safety metadata are runtime confirmation/authorization
  constraints, never disablement reasons. See `PREMERGE-GATE.md` and the
  generated ledger's `pre_merge_gate` object.

- [x] Existing PR #4301 retargeted through the GitHub API to
  `fm/cli-reverse-etl-destination-r1`; API read-back reported that exact base.
- [x] Foundation SHA `c6f03c937c1f4e516d339b48e8c2143726179fdf` merged as
  `16c047eaf` without history rewriting.
- [x] CI Verify failure root-caused to Lever source-map regeneration, then
  reproduced locally.
- [x] Lever focused repair: `go test -timeout 20m ./cmd/connectorgen -run
  '^TestLeverHiringAPISurfaceOperationLedger$'` and Batch 9 source integrity
  check pass.
- [x] Generated and validated the 30-row seven-surface ledger and human
  summary; every one of 1,591 typed actions has a reverse-ETL eligibility and
  a direct-write CLI disposition.
- [x] Added only faithful connector-local source/destination declarations and
  per-increment static evidence; generic App/CLI dispatch remains explicitly
  pending the foundation integration.
- [ ] Before final push: fetch and merge the latest foundation branch, prove
  its exact SHA is an ancestor, and exercise the installed App/CLI dispatch
  route without credentials.

### First five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check brex zoho-books testrail amplitude posthog` failed before declarations with `brex: source transport declaration missing`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Brex, Zoho Books, TestRail, Amplitude, and PostHog; generated
  `SEVEN-SURFACE-LEDGER.json` and `SEVEN-SURFACE-SUMMARY.md` for all 30
  connectors.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check brex zoho-books testrail amplitude posthog`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### First five-connector typed-write CLI increment

- [x] Red: the prior generated ledger contained 1,425
  `declaration-pending-cli-binding` typed actions; a static destination proof
  does not make an action CLI-reachable.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli brex zoho-books testrail amplitude posthog`
  generated only existing-action command contracts. Binding is fail-closed on
  the exact singular `covered_by.write` action, equal method, canonical path,
  declared record placeholders, and declared query/config inputs.
- [x] `go run .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-promotable-writes.go`
  — real engine schema gate: 1,591 actions, 1,591 promotable, 0 foundation
  gaps.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check brex zoho-books testrail amplitude posthog`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Second five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check metabase dbt looker mode dremio` failed before declarations with `metabase: source transport declaration missing`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Metabase, dbt Cloud, Looker, Mode, and Dremio; the generated all-30
  ledger records all typed actions' eligibility and CLI-binding disposition.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check metabase dbt looker mode dremio`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### Second five-connector typed-write CLI increment

- [x] Red: dbt Cloud and Dremio had 24 missing typed-write CLI bindings;
  Metabase, Looker, and Mode expose no existing typed actions and therefore
  cannot receive an invented command.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli metabase dbt looker mode dremio`
  regenerated their connector-owned surfaces under the same exact-binding
  rule; zero-action bundles remain command-free.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check metabase dbt looker mode dremio`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Third five-connector typed-write CLI increment

- [x] Red: 108 Coda, ClickUp, Calendly, Greenhouse, and Lever typed actions
  lacked a complete CLI disposition; safety or privilege is not an eligibility
  exclusion.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli coda clickup-api calendly greenhouse lever-hiring`
  gives every existing action a connector command. Exact-route failures remain
  partial (Coda/ClickUp), while exact contracts are implemented (Calendly 6,
  Greenhouse 58, Lever 14).
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check coda clickup-api calendly greenhouse lever-hiring`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Third five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check coda clickup-api calendly greenhouse lever-hiring` failed before declarations with `coda: source transport declaration missing`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Coda, ClickUp, Calendly, Greenhouse, and Lever Hiring; each existing
  typed action has an eligibility and direct-CLI-binding disposition in the
  regenerated 30-row ledger.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check coda clickup-api calendly greenhouse lever-hiring`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### Fourth five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check ashby workable recruitee hibob factorial` failed before declarations with `ashby: source transport declaration missing`; initial generation also failed real `connectorgen validate` because the closed mapper rejects `applicationId`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Ashby, Workable, Recruitee, HiBob, and Factorial. The ledger records the
  concrete mapper refusal at `internal/connectors/sync_transport.go:673` and
  the minimal foundation change for every case-preserving input, while valid
  lowercase mappings remain declared.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check ashby workable recruitee hibob factorial`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### Fourth five-connector typed-write CLI increment

- [x] Red: real `TestEveryImplementedCommandPassesRuntimePreflight` rejected
  Ashby's `create survey submission apply`: runner line 484 requires the
  native connector to delegate the closed structured-JSON record-schema gate.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli ashby workable recruitee hibob factorial`
  records that command partial with the exact refusal/minimal bridge change;
  it gives every existing Ashby and Workable action a command and leaves the
  zero-action bundles free of invented CLI actions.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check ashby workable recruitee hibob factorial`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Fifth five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check datadog pagerduty auth0 okta firehydrant` failed before declarations with `datadog: source transport declaration missing`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Datadog, PagerDuty, Auth0, Okta, and FireHydrant. Every typed action is
  listed in the generated ledger with exact eligibility, CLI-binding, and
  case-preserving-mapping foundation-gap status where applicable.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check datadog pagerduty auth0 okta firehydrant`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### Fifth five-connector typed-write CLI increment

- [x] Red: 708 typed Datadog, Auth0, Okta, and FireHydrant actions remained
  CLI-unbound; PagerDuty has no existing typed action and must not receive an
  invented command.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli datadog pagerduty auth0 okta firehydrant`
  gives every existing action a connector command under strict exact binding
  (implemented/partial: Datadog 13/14, Auth0 6/2, Okta 105/324,
  FireHydrant 0/244).
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check datadog pagerduty auth0 okta firehydrant`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Sixth five-connector declaration increment

- [x] Red: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check adobe-commerce-magento commercetools recharge docuseal eventbrite` failed before declarations with `adobe-commerce-magento: source transport declaration missing`.
- [x] Green: wrote only connector-owned `sync_transport.json` declarations
  for Adobe Commerce, commercetools, Recharge, DocuSeal, and Eventbrite;
  source-only bundles retain no invented destination action.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check adobe-commerce-magento commercetools recharge docuseal eventbrite`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.

### Sixth five-connector typed-write CLI increment

- [x] Red: Adobe Commerce and DocuSeal's ten existing typed actions lacked
  reconciled commands; Commercetools, Recharge, and Eventbrite are source-only
  because their bundles have no typed action.
- [x] Green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --write-cli adobe-commerce-magento commercetools recharge docuseal eventbrite`
  creates all ten action commands (Adobe Commerce 0/4 implemented/partial;
  DocuSeal 6/0), without inventing source-only commands.
- [x] The all-30 ledger reports 1,591 typed actions, 342 implemented commands,
  1,249 precise partial commands, and zero
  `declaration-pending-cli-binding` actions.
- [x] `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/reconcile-seven-surfaces.mjs --check adobe-commerce-magento commercetools recharge docuseal eventbrite`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552
  connector(s) checked, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connector(s)
  scanned, 0 changes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

## Source-first map checks

- [x] Red: initial maps rejected because their locks lacked complete provider
  source evidence, total counts, and independent coverage basis.
- [x] Green: Batch 8 integrity map check passes.
- [x] Green: Batch 9 integrity map check passes.
- [x] Green: Batch 10 integrity map check passes.
- [x] Every direct write is primary class `direct_write`; reverse ETL is only
  the `generic-typed-destination-executor` eligibility attribute.
- [x] All 30 old/new counts and source bases are recorded in
  `SOURCE-SURFACE-REPORT.md`.
- [x] Red/green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-reverse-etl-readiness.mjs --check`
  confirms direct-write action coverage without claiming a destination before
  #4303 (26 mapped connectors; 6,311 writes; 1,419 action-backed; 4,892 pending
  connector-local typed-action authoring; 4 unavailable/dynamic sources), and
  confirms all 1,419 action-backed source/action bindings in
  `REVERSE-ETL-TYPED-ACTION-INVENTORY.json` remain `not-declared` under the
  uniform #4303 foundation gap.

## Completed generated/bundle checks

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  — `552 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen surface-sync --check`
  — `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected
  across 0 connector(s)`.
- [x] After the reverse-ETL readiness artifacts: reran
  `go run ./cmd/connectorgen validate internal/connectors/defs` (552 checked,
  0 findings) and `go run ./cmd/connectorgen surface-sync --check` (exit 0;
  no surface drift reported).
- [x] `git diff --check`.
- [x] `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/conformance ./internal/connectors/commandrunner`.
- [x] `go test -timeout 20m ./internal/cli`.
- [x] `go build ./cmd/pm` and `go vet ./...`.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, and
  `make release-workflow-check`.
- [x] `go run ./cmd/connectorgen boundary . --json`.

## Pending final gates

- [x] Rebased to `origin/main` at `51dd6d468` (including PR #4297), then
  reran all three declaration integrity checks, connector validation,
  `surface-sync --check`, targeted Go tests, `go build`, lint, and the agent
  contract check. No map retains the superseded HEAD/pagination engine-gap IDs.
- [x] Full uncached `go test -timeout 20m ./internal/cli` after the supported
  nine-root-transcript refresh — passed in 779.762s. The focused
  `TestGoldenTranscripts` red/green repair passed first.
- [x] Built `/tmp/pm-4292-reachability-probe`, initialized a credential-free
  project, and exercised the real connector CLI: Brex `update vendor apply`
  stops at `error: missing --credential`; Zoho `update contact person 2 apply`
  returns its exact singular-`covered_by.write` partial blocker.
- [x] Fetched `origin/fm/cli-reverse-etl-destination-r1` at
  `c6f03c937c1f4e516d339b48e8c2143726179fdf`; it is an ancestor of this
  branch. Its installed `pm etl` manual exposes only the existing closed
  GitHub/PostgreSQL transport routes, so application-level declarative
  destination dispatch remains correctly pending the foundation lane.
- [x] Final map/ledger gates: `verify-parity-maps.mjs` for batches 8, 9, and
  10; `reconcile-seven-surfaces.mjs --check` for all 30; real promotability
  audit; `connectorgen validate`; and `surface-sync --check` all passed.
- [x] Final repository gates: `go test -timeout 20m` for connectorgen,
  conformance, and commandrunner; `go vet ./...`; `go build ./cmd/pm`; and
  individual `make tidy-check`, `lint`, `docs-check-no-build`,
  `smoke-no-build`, `agent-contract-check`, and `release-workflow-check`
  passed. Detached `go run ./cmd/connectorgen boundary . --json` completed
  with zero findings (capture: `/tmp/pm-4292-connector-boundary.json`).
- [ ] Refresh the foundation head once its App/CLI dispatch lands, exercise
  that installed route, then push and update existing PR #4301.
