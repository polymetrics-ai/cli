# Issue #4292 — TDD ledger

## Reconciliation red/green evidence — 2026-08-20

### Red

- Verify CI run `32281460555` failed in
  `TestLeverHiringAPISurfaceOperationLedger`: regenerated blocked rows omitted
  its required named dependency. Reproducing the focused test after the first
  metadata-only repair exposed the second source-normalization defect: the
  document's no-leading-slash `GET /v1/eeo/responses` route had been lost.

### Green

- `generate-parity-maps.mjs` now emits a concrete
  `Named dependency: connector-local typed operation declaration ...` for a
  source-derived blocked row.
- `extract-source-operations.go` records the Lever rendered-reference
  normalization exceptions: it excludes prose-only `GET /profile_forms` and
  includes the documented EEO route with a precise explanation of the source's
  missing leading slash. No request/body schema is created by this exception.
- `node .../generate-parity-maps.mjs lever-hiring` regenerated only the owned
  Lever artifacts. `go test -timeout 20m ./cmd/connectorgen -run
  '^TestLeverHiringAPISurfaceOperationLedger$'` and
  `node .../verify-parity-maps.mjs 9` pass.

### Pending next red/green slice

- Add a deterministic seven-surface declaration verifier before writing the
  connector-local source/destination declarations. It must reject absent
  streams/actions, non-`input_fields` destination mapping, and a false claim
  that generic App/CLI dispatch is deployed before the newer #4304 foundation
  head is merged and exercised.

### Seven-surface declaration proof — first increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check brex zoho-books
  testrail amplitude posthog` failed with `brex: source transport declaration
  missing`. This asserts that a source inventory alone cannot be presented as
  a transport declaration.
- **Green:** the same trace generated connector-owned source declarations for
  all five bundles. It selected one non-destructive, exact-schema typed action
  as an initial destination proof where one exists, listed every other
  record-driven typed action as eligible, and preserved a per-action
  multiplicity dependency when the one-action-per-mode model cannot select it.
  PostHog has no existing typed action and therefore has no invented
  destination declaration.
- `node traces/reconcile-seven-surfaces.mjs --check brex zoho-books testrail
  amplitude posthog`, `go run ./cmd/connectorgen validate
  internal/connectors/defs`, and `go run ./cmd/connectorgen surface-sync
  --check` pass. The generated 30-row ledger keeps generic App/CLI dispatch
  explicitly pending foundation integration.

### Seven-surface declaration proof — second increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check metabase dbt
  looker mode dremio` failed with `metabase: source transport declaration
  missing` before any declaration was written.
- **Green:** generated connector-owned source declarations for Metabase, dbt
  Cloud, Looker, Mode, and Dremio. dbt Cloud and Dremio each have existing
  exact-record actions, so their destination declarations list every such
  action and select only one initial `full_append` proof. Metabase, Looker,
  and Mode have no existing typed actions; no destination contract or request
  schema was invented.
- The focused checker, `connectorgen validate`, and `surface-sync --check`
  pass. Generic App/CLI dispatch and multi-action selection remain foundation
  dependencies, recorded rather than claimed as deployed.

### Seven-surface declaration proof — third increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check coda
  clickup-api calendly greenhouse lever-hiring` failed with `coda: source
  transport declaration missing` before connector-owned declarations existed.
- **Green:** generated source declarations for Coda, ClickUp, Calendly,
  Greenhouse, and Lever Hiring. Existing exact actions in ClickUp, Calendly,
  Greenhouse, and Lever are all named in `eligible_actions`; Coda's semantic
  exclusions name missing source-record fields, without manufacturing a
  destination. The per-action ledger also marks absent direct CLI bindings as
  `declaration-pending-cli-binding`, preserving the outstanding reachability
  work rather than treating it as a safety exclusion.
- The focused checker, `connectorgen validate`, and `surface-sync --check`
  pass.

### Seven-surface declaration proof — fourth increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check ashby workable
  recruitee hibob factorial` failed with `ashby: source transport declaration
  missing`. The first generated Ashby declaration then exercised the real
  loader and failed `connectorgen validate`: `applicationId` violates the
  closed mapping identifier rule.
- **Green:** the reconciliation trace now gives every action with a
  case-preserving provider input an explicit `foundation-gap` disposition:
  `internal/connectors/sync_transport.go:673` refuses it, and the minimal
  change is allowing case-preserving concrete `input_fields` names so action
  property names remain exact. It does not rewrite or lowercase the provider
  action input. Ashby's valid lowercase mappings are still declared; Workable
  declares every valid record-driven action; Recruitee, HiBob, and Factorial
  do not invent destinations without typed actions.
- The regenerated declarations pass the focused checker, real
  `connectorgen validate`, and `surface-sync --check`.

### Seven-surface declaration proof — fifth increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check datadog
  pagerduty auth0 okta firehydrant` failed with `datadog: source transport
  declaration missing`.
- **Green:** generated source declarations for all five bundles, declarative
  destination proofs for Datadog, Auth0, Okta, and FireHydrant, and no
  invented PagerDuty destination because it has no typed action. The Okta
  case-preserving inputs are recorded as the same exact `sync_transport.go:673`
  foundation gap rather than silently relabelled as semantic exclusions.
- The focused checker, real `connectorgen validate`, and `surface-sync --check`
  pass.

### Seven-surface declaration proof — sixth increment

- **Red:** `node traces/reconcile-seven-surfaces.mjs --check
  adobe-commerce-magento commercetools recharge docuseal eventbrite` failed
  with `adobe-commerce-magento: source transport declaration missing`.
- **Green:** generated the final five connector-owned source declarations and
  exact-record destination proofs for Adobe Commerce and DocuSeal. The three
  bundles without typed actions receive source-only declarations; no generic
  request or fabricated destination action was introduced. This completes the
  all-30 declaration pass, while the machine ledger remains explicit that
  direct CLI action binding and foundation dispatch/multiplicity work are not
  complete.
- The focused checker, real `connectorgen validate`, and `surface-sync --check`
  pass.

### Typed-write CLI proof — first increment

- **Red:** the all-30 ledger exposed 1,425 typed actions with
  `declaration-pending-cli-binding`; a destination declaration is not an
  executable CLI surface. The first Zoho attempt also showed that its legacy
  `covered_by.writes` array cannot be treated as an exact operation binding.
- **Green:** `reconcile-seven-surfaces.mjs --write-cli brex zoho-books
  testrail amplitude posthog` materializes connector-owned command surfaces
  from existing typed actions only. Its route join accepts exactly one
  `api_surface.covered_by.write` equal to the action name, requires method and
  canonical path equality, promotes only declared `{{ record.<field> }}` path
  placeholders, and treats URL query templates as typed query/config flags.
  Zero/multiple matches, method/path disagreement, and unsupported placeholder
  syntax remain partial with a precise declared reason; no fuzzy match or
  array-order selection occurs. Query config inputs are preserved as typed
  flags and are partial only for the concrete
  `internal/connectors/commandrunner/runner.go:1565` closed-record override
  limitation.
- `go run traces/audit-promotable-writes.go` calls the real
  `engine.ValidatePromotableRecordSchema` gate for all 1,591 actions: 1,591
  promotable, 0 foundation gaps. The first five action surfaces then pass
  `node traces/reconcile-seven-surfaces.mjs --check brex zoho-books testrail
  amplitude posthog`, `connectorgen validate`, `surface-sync --check`, and
  `go test -timeout 20m ./internal/connectors/commandrunner -run
  '^TestEveryImplementedCommandPassesRuntimePreflight$'`.

### Typed-write CLI proof — second increment

- **Red:** the second five bundles still carried 24 absent typed-action CLI
  bindings (dbt Cloud and Dremio); the zero-action Metabase, Looker, and Mode
  rows must not grow invented commands.
- **Green:** materialized only dbt Cloud and Dremio actions in their existing
  connector-owned surfaces. The exact route join and partial reasons are the
  same closed rule as the first increment; the three source-only bundles gain
  no action command. The focused trace, `connectorgen validate`,
  `surface-sync --check`, and runtime-preflight sweep pass.

### Typed-write CLI proof — third increment

- **Red:** Coda, ClickUp, Calendly, Greenhouse, and Lever still had no complete
  action CLI disposition under the reconciled ledger. The substantial Coda and
  ClickUp action sets could not be silently omitted because their actions are
  privileged or uncommon.
- **Green:** all 108 existing actions received a connector command. Coda and
  ClickUp retain partial commands where no singular exact provider route meets
  the declared binding rule; Calendly has six implemented and two precise
  partials; Greenhouse and Lever have 58 and 14 implemented commands,
  respectively. The trace, `connectorgen validate`, `surface-sync --check`,
  and the real all-implemented runtime-preflight sweep pass.

### Typed-write CLI proof — fourth increment

- **Red:** the initial runtime-preflight sweep rejected Ashby's generated
  `create survey submission apply`: its declared `submittedValues` JSON flag
  is closed and typed, but the registered native connector does not expose the
  declarative record-schema preflight interface required at
  `internal/connectors/commandrunner/runner.go:484`.
- **Green:** the reconciliation generator preserves that command as partial
  with the exact refusal and minimal native-bridge delegation change, then
  reruns deterministically without selecting the wrong duplicate partial
  command. Ashby has 93 implemented/5 partial commands and Workable's 38
  actions all receive exact-route partial commands; source-only Recruitee,
  HiBob, and Factorial do not gain invented actions. The focused trace,
  `connectorgen validate`, `surface-sync --check`, and runtime-preflight sweep
  pass.

### Typed-write CLI proof — fifth increment

- **Red:** the 708 unbound typed actions concentrated in Datadog, Auth0, Okta,
  and FireHydrant could not remain ledger-only. PagerDuty's zero-action bundle
  is not permission to invent a command.
- **Green:** generated command contracts for all 708 actions. Exact matches
  are implemented where the closed route/input contract reaches runtime
  preflight; every remaining action is partial with its exact singular-route,
  config-override, record-shape, or case-preserving-input blocker. Datadog has
  13 implemented/14 partial, Auth0 6/2, Okta 105/324, and FireHydrant 0/244.
  The focused trace, bundle validation, surface synchronization, and uncached
  all-implemented runtime-preflight sweep pass.

## Red

- The captain's `SOURCE-LOCK-DEFECT.md` established the initial red state:
  maps used landing-page pins, omitted `counts.total`, bounded their result by
  `api_surface.json`, and reported self-referential `declared_percent`.
- The initial integrity assertion correctly rejected the old Batch 8 source
  locks as incomplete source evidence. It also caught the pre-fix
  per-method-count implementation error while the new source extractor and
  cross-artifact verifier were being introduced.
- A source map that put a typed write under `reverse_etl` is red: write
  endpoints are `direct_write`; reverse-ETL is only the destination-executor
  eligibility attribute.

## Green

- Added `extract-source-operations.go`, which pins every provider document's
  exact byte count and SHA-256, then extracts REST method/path operations from
  complete OpenAPI/Swagger documents or complete official rendered references.
- Regenerated all mapped `api_surface.json` files from that source inventory,
  retaining old bindings only when they resolve to a pinned operation. The
  generated report records prior and refreshed counts for every connector.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 8`
  passed for all Batch 8 connectors.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 9`
  passed for all Batch 9 connectors.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 10`
  passed for all Batch 10 connectors.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed with
  `552 connector(s) checked, 0 findings`.
- `go run ./cmd/connectorgen surface-sync --check` passed with 552 scanned and
  zero fields changed.

## Reverse-ETL readiness freeze — Red

- Issue #4303's verified current state is red for every connector: the only
  destination factory is GitHub issue-label-bound. A typed direct-write action
  is therefore not evidence of an executable reverse-ETL destination.
- Any readiness artifact that classifies a row as primary `reverse_etl`, says an
  action-backed write is destination-enabled, or adds a `transport_binding`
  fails the captain's safety boundary.

## Reverse-ETL readiness freeze — Green

- Added `traces/audit-reverse-etl-readiness.mjs`, which generates and checks a
  deterministic all-30 readiness report from the source-first ledgers. It
  rejects a primary `reverse_etl` row, a missing uniform #4303 gap, an enabled
  direct write without its named typed action, or an invented
  `transport_binding`.
- The same trace generates `REVERSE-ETL-TYPED-ACTION-INVENTORY.json`, preserving
  all 1,419 action-backed source ID/route/action-ID bindings, with a pointer to
  each disposition ledger for the exact source location, as
  pre-foundation inputs. The artifact explicitly sets the entire inventory to
  `not-declared`; it applies the connector-owned evidence, bindings, acknowledgement,
  mode-strategy, and product-safety work still required after #4303 uniformly.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-reverse-etl-readiness.mjs --check`
  passed: 26 source-enumerable connectors contain 6,311 direct-write
  operations; 1,419 rows are already action-backed, 4,892 need connector-local
  typed-action authoring, and four unavailable/dynamic provider surfaces remain
  explicitly non-enumerable.
- The report records Zoom's captain-directed 204-action future destination
  cohort as a planning target, not an invented source count or declaration.

## Refactor / review

- The source lock carries `counts.total`, per-method and per-kind counts,
  source-document pins, and coverage basis. Dispositions use
  `operations_found`; no generated summary has `declared_percent`.
- TestRail, Eventbrite, and Greenhouse are explicitly skipped with browser
  evidence and `counts.total: null`; Adobe Commerce is explicitly
  `dynamic-instance-dependent` with a pinned official rendered reference and
  no fabricated total.
- Every `direct_write` has the locked
  `generic-typed-destination-executor` reverse-ETL eligibility gap; no row is
  primarily classed `reverse_etl` and no `transport_binding` action exists.
