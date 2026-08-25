# Issue #4292 — TDD ledger

## Reconciliation red/green evidence — 2026-08-20

### Declared transport batch regression — 2026-08-24

- **Red:** `GOFLAGS=-p=3 go test -timeout 20m ./internal/connectors/certify`
  fails `TestCertificationDeclaredTransportPairResolvesAndExecutes` before
  provider I/O: the #4292 reconciliation generator emits a
  `declarative_typed_destination` `source_bindings` entry without the newly
  required sealed `batch` declaration. Clean `origin/main` passes the same
  suite, proving this is branch-generated declaration drift rather than a
  pre-existing certification failure.
- **Green target:** regenerate every assigned connector transport through the
  canonical reconciliation trace with a closed
  `batch: {"disposition":"per_record","max_records":1}` for each exact
  record-driven binding; retain the certification assertion unchanged and
  prove the focused suite plus generator/surface checks.
- **Second red:** after the batch repair, the unchanged certification test
  reached the next closed-contract check: Adobe Commerce's `update_product`
  has no provider-documented `idempotency_key_header`. The same inspection
  found no such header on any action selected by these thirty generated
  destination declarations. These actions remain direct-write CLI reachable;
  they cannot be falsely declared as replay-safe reverse-ETL destinations.
- **Second green target:** make the canonical candidate selector require the
  provider-documented idempotency header before it emits a reusable typed
  destination binding, then regenerate all assigned declarations and rerun
  the unchanged certification suite.
- **Green:** the canonical trace now seals `per_record`/`max_records: 1` on
  every emitted binding and requires a documented `idempotency_key_header`
  before it declares a reusable destination. Regeneration removed the
  unsupported destination claims while preserving every generated direct-write
  command. The unchanged `GOFLAGS=-p=3 go test -timeout 20m
  ./internal/connectors/certify` suite passes in 38.273s.

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

### Typed-write CLI proof — sixth increment

- **Red:** the final five bundles still left Adobe Commerce's four and
  DocuSeal's six existing typed actions without a reconciled command; the
  Commercetools, Recharge, and Eventbrite bundles have no existing typed
  actions and must remain source-only.
- **Green:** all ten existing actions receive connector command contracts:
  Adobe Commerce's four preserve exact partial reasons and DocuSeal's six pass
  runtime preflight as implemented. The all-30 ledger now has 1,591 typed
  actions and zero `declaration-pending-cli-binding` rows (342 implemented,
  1,249 precise partials). The focused trace, bundle validation, surface sync,
  and uncached runtime-preflight sweep pass.

### CLI manual snapshot proof

- **Red:** the full `internal/cli` suite exposed stale root-help golden
  transcripts after the generated connector namespaces became visible in the
  runtime manual. The failure is a snapshot parity failure, not a connector
  dispatch failure.
- **Green:** the test-supported
  `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1` path refreshed only the nine root
  help variants; the focused transcript test and a full uncached
  `go test -timeout 20m ./internal/cli` rerun pass. The installed binary also
  proves an implemented Brex command reaches `missing --credential` in an
  initialized credential-free project, while a no-singular-binding Zoho action
  returns its exact declared partial block reason.

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

### Ashby declaration regression from generated command expansion — 2026-08-24

- **Red:** CI's complete `make test` run and the focused
  `TestOperationClassificationsMatchAshbySemantics` /
  `TestCustomFieldValueCommandsArePartial` failed: the migrated definition
  replaced the named multipart foundation with a generic missing-action reason
  and emitted scalar `fieldValue` flags as implemented despite the provider's
  documented boolean/number/string/array/object/null union. The prior fixed
  blocked-row total (34) also became stale after 19 declaration-owned actions
  replaced former generic blocked rows; the two true named foundations remain
  mandatory.
- **Green:** the connector definition again names the multipart and
  side-effect foundations (`ashby-application-form-typed-multipart-foundation`
  and `ashby-referral-form-info-side-effect-foundation`); each generated custom
  field command remains CLI-declared but partial without a lossy scalar union
  flag. No shared runtime branch or connector-name special case is used.

### Shared source-factory evidence integration regression — 2026-08-24

- **Red:** `TestDefinitionTransportFactoriesSelectDeclaredEvidence` passed on
  merge base `060bb7864e` (2.520s) but failed on this branch because the 30
  connector-owned `sync_transport.json` source declarations correctly added
  their own static evidence to the reusable factory. The old assertion treated
  GitHub's evidence as the arbitrary primary entry, even though
  `AcceptedSourceEvidences` is the installed multi-declaration contract.
- **Green:** the test now proves the stronger runtime property: GitHub's exact
  declaration evidence must be accepted by the shared source factory, whether
  it is primary or an additional declaration-owned evidence reference. No
  connector name branch, engine behavior, or source declaration was changed.

### Root help transcript parity after Batch 8–10 command publication — 2026-08-25

- **Red:** CI job `32763389058` failed all nine root manual/help transcript
  variants because the root command correctly listed the newly published Batch
  8–10 connector command namespaces while
  `internal/cli/testdata/golden_transcripts.json` still described the older
  catalog. The same selected transcript test reproduced locally.
- **Green:** regenerated exactly the nine affected root transcript fixtures
  (`root_bare_manual`, all help/manual variants, and their JSON/equal/space
  forms) using the in-repository golden writer. The selected test now passes;
  no command implementation or runtime parsing changed.

### Connector manual catalog parity after Batch 8–10 command publication — 2026-08-25

- **Red:** CI job `32766058702` passed all Go tests, including the repaired
  root transcripts, then failed `./pm docs validate --connectors-dir
  docs/connectors` because `docs/connectors/catalog/all-connectors.json` and
  the affected connector manuals did not yet reflect the Batch 8–10 command
  namespaces. The validation failure reproduced locally.
- **Green:** regenerated `docs/cli` and `docs/connectors` with
  `pm docs generate --dir docs/cli --connectors-dir docs/connectors`, then
  reran `pm docs validate --connectors-dir docs/connectors` successfully.
  This is generated documentation parity only; command behavior is unchanged.

### Source-lock projection foundation — 2026-08-25

- **Red:** PR #4301 current head `feb0ac324` fails
  `connectorgen validate` with 30 `source_projection` findings, while the
  preserved `origin/main` tree validates clean (553 connectors, 0 findings).
  The difference is the branch's 30 source-lock imports: main has no such
  locks to project. The shared validator's `checkSourceProjection`
  (`cmd/connectorgen/sourceprojection.go:2244-2268`) requires a separate
  canonical `*-operation-descriptor.json`; it has no projector from v3
  `source_documents`. Separately, `validateSourceImportREST`
  (`cmd/connectorgen/sourceimport.go:809-813`) rejects imported OpenAPI
  document metadata unless it is represented in the aggregate inventory.
- **Foundation gap:** this is not a connector-specific exception or a reason
  to fabricate descriptors, pins, or operation counts. The required shared
  component is the v3 source-lock-to-canonical-descriptor projector (including
  aggregate OpenAPI inventory derivation) that preserves captured-document
  evidence. Declaration-first mapping continues against the checked-in
  provider evidence; source certification remains blocked until that component
  lands.

### Declaration-first direct-write cohort 01 — 2026-08-25

- **Red:** no checked-in artifact named the source identity, current typed
  action CLI path, implementation lane, and exact deferred component for a
  bounded Batch 8 cohort. The existing aggregate ledgers prove source/action
  counts but do not make promotion mechanical per provider operation.
- **Green target:** a deterministic generator will derive the first five
  Batch 8 connectors (Brex, Zoho Books, TestRail, Amplitude, and PostHog) from
  their committed crosswalks and dispositions. It must preserve existing typed
  action CLI paths verbatim, emit no path for an unmaterialized operation, and
  name `source-lock-projection-gap` as the shared source-certification blocker.
  TestRail must remain unavailable rather than becoming a zero-operation map.
  The initial red run also proved that a missing `writes.json` is a meaningful
  empty action set (PostHog), not a malformed connector or a license to infer
  actions; the generator now makes that state explicit.
- **Green:** `generate-declaration-first-cohort.mjs --check` now verifies the
  checked-in machine map and human summary for 1,861 direct-write source rows:
  Brex 49 (14 existing-schema CLI-bound), Zoho Books 579 (562), TestRail
  unavailable, Amplitude 99 (12), and PostHog 1,134 (0). Every enumerable row
  cites its exact source lock ID/location, reports only an already-declared
  typed-action CLI path, records its connector-local lane, and carries the
  same source-certification-only `source-lock-projection-gap`. The 1,273
  remaining rows have no fabricated path or request contract.

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
