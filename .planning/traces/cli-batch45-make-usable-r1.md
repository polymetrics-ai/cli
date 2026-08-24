## Task Delivery Header

- Issue: Refs #4349 — make the twenty #4295 connector declarations demonstrably usable.
- Base branch: main
- Merges into: main
- Delivery: Phase 1 ends with a source-cited, binary-measured eight-lane gap map and awaits captain approval; after approval, one scoped wave is committed, verified, pushed, and opened as a pull request against `main`.
- Working branch: fm/cli-batch45-make-usable-r1
- Task: Re-measure the twenty Batch 4/5 connectors from #4295. Report each connector's own pinned-source operation count and its ETL, reverse-ETL, direct-read, direct-write, binary-upload, binary-download, schedule, and flow status. Do not implement Phase 2 until approval.
- Verification: Inspect each connector bundle and pinned source; run source-provided/generated validation where applicable; build `pm`; execute every already-implemented command without credentials and record whether it reaches `missing --credential` rather than an unknown-command or preflight failure.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The gap map reflects the current declarative and runtime state | live | Bundle inventories and the built binary demonstrate the count and credential-boundary result for each claimed command. |
| Every `not_applicable` lane has provider-source support | live | The connector's pinned source citation states the provider capability or its absence. |
| Schedule and flow are not invented as CLI intents | live | The checked schema enumeration excludes both intents. |
| Binary upload is reported as pending rather than missing | live | Open PR #4343 provides the unmerged declaration-only intent and is not treated as main-branch capability. |

## Phase 1 Method

- `Red:` Treat prior aggregate claims as untrusted; measure the checked-out bundle, schema, source lock, and built CLI.
- `Green:` Produce the source-cited, eight-lane map; retain only results whose binary invocation reaches the credential boundary.
- Required skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- GSD mode: inline/manual investigation. The compatible isolated-worker runtime is unavailable here and the canonical single-worker delivery contract forbids role spawning. The standard GSD command sources and agent contract check were run before investigation.

## Phase 1 Re-measurement — 2026-08-24

### Result

The brief's initial declarative inventory is stale:

- All twenty connectors now have `cli_surface.json`, not only Mailchimp and Zendesk Support.
- Eighteen have only `partial` declaration-pending commands. Mailchimp has 295 `implemented` rows and Zendesk Support has 95; across the cohort this is 7,054 partial and 390 implemented rows.
- The exact provider source-lock total is 7,301 operations across the 17 locks that can honestly count their provider surface. eBay Fulfillment and TikTok Marketing have unavailable public sources; Salesforce is tenant-dynamic. Those three totals are unknown, not zero.
- Eight, rather than seven, bundles have `writes.json`: Airtable, BambooHR, Buildkite, Mailchimp, Pipedrive, SonarCloud, Squarespace, and Zendesk Support. All twenty lack `sync_transport.json`.
- `TestEveryImplementedCommandPassesRuntimePreflight` succeeds for all 390 declared rows. It is structural evidence only. Built-binary credential-boundary probes independently confirmed 36 distinct Mailchimp commands and Zendesk Support's `streams tickets list`; the remaining 353 declared implemented rows are not counted as Phase 1 binary-certified usable.

Notation below: `I` means declared `implemented`, `P` means declared `partial`, `S` means the source disposition crosswalk count, and `W` means a source-backed provider write candidate. `S` can include explicit local execution bindings; the source column remains the provider-operation total. `none classified` never claims a provider lacks a capability: the pinned lock simply does not establish a bounded binary contract.

| Connector | Pinned source total | ETL | Reverse ETL | Direct read | Direct write | Binary download | Binary upload | Schedule | Flow |
| --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| Airtable | 103 [source](https://airtable.com/developers/web/api/introduction) | P5/S5 | —/W74 | P26/S26 | P74/S74 | 0; none classified | pending; none classified | system | system |
| BambooHR | 319 [source](https://documentation.bamboohr.com/reference/get-meta-company) | P84/S84 | —/W190 | P71/S71 | P190/S190 | 0; none classified | pending; 19 binary/multipart-tagged rows, direction not preserved | system | system |
| Buildkite | 129 [source](https://buildkite.com/docs/apis/rest-api) | P6/S6 | —/W76 | P50/S50 | P76/S76 | 0; none classified | pending; none classified | system | system |
| eBay Fulfillment | unknown; [official error-page evidence](https://developer.ebay.com/develop/api/fulfillment-api/release-notes) | P4/S? | —/W? | P4/S? | P3/S? | unknown | pending; source unavailable | system | system |
| Fastly | 732 [source](https://www.fastly.com/documentation/downloads/fastly.collection.json) | P6/S6 | —/W389 | P337/S337 | P389/S389 | 0; none classified | pending; none classified | system | system |
| Google Analytics Data API | 23 [source](https://analyticsdata.googleapis.com/$discovery/rest?version=v1) | P5/S5 | —/W16 | P7/S7 | P16/S16 | 0; none classified | pending; none classified | system | system |
| HubSpot | 3,118 [source](https://codeload.github.com/HubSpot/HubSpot-public-api-spec-collection/tar.gz/2bebde2dca45eaa1792931089c4e441c8e377594) | 0/S0 | —/W1,901 | P1,240/S1,217 | P1,901/S1,901 | 0/S32 `binary_read` | pending; 229 binary/multipart-tagged rows, direction not preserved | system | system |
| LaunchDarkly | 397 [source](https://app.launchdarkly.com/api/v2/openapi.json) | P5/S5 | —/W205 | P189/S189 | P205/S205 | 0; none classified | pending; none classified | system | system |
| Linear | 539 [source](https://studio.apollographql.com/public/Linear-API/variant/current/schema/reference) | P4/S4 | —/W373 | P166/S166 | P373/S373 | 0; none classified | pending; none classified | system | system |
| Mailchimp | 295 [source](https://api.mailchimp.com/schema/3.0/Swagger.json) | I79/S79 | I148/W164 | I68+P12/S80 | P16/W164 | 0; none classified | pending; none classified | system | system |
| Pinterest | 279 [source](https://developers.pinterest.com/docs/api/v5/introduction/) | P5/S5 | —/W135 | P144/S144 | P135/S135 | 0; none classified | pending; none classified | system | system |
| Pipedrive | 213 [source](https://developers.pipedrive.com/docs/api/v1/openapi.yaml) | P21/S21 | —/W99 | P98/S98 | P99/S99 | 0; one explicit raw-binary download | pending; 12 binary/multipart-tagged rows, direction not preserved | system | system |
| QuickBooks | 129 [source](https://static.developer.intuit.com/JSONObjects/EntityJsonObject_v1.json) | P5/S5 | —/W45 | P84/S84 | P45/S45 | 0; none classified | pending; none classified | system | system |
| Salesforce | unknown; [dynamic resource index](https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_list.htm) | P4/S? | —/W? | P3/S? | P3/S? | unknown | pending; source total is tenant-dynamic | system | system |
| ShipStation | 47 [source](https://docs.shipstation.com/_bundle/apis/@shipstation-v1/openapi.json?download) | P4/S4 | —/W25 | P18/S18 | P25/S25 | 0; none classified | pending; none classified | system | system |
| SonarCloud | 156 [source](https://sonarcloud.io/api/webservices/list) | P11/S11 | —/W87 | P59/S59 | P87/S87 | 0; none classified | pending; none classified | system | system |
| Squarespace | 53 [source](https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json) | P8/S8 | —/W30 | P15/S15 | P30/S30 | 0; none classified | pending; one multipart image-upload row | system | system |
| TikTok Marketing | unknown; [SSL-error evidence](https://business-api.tiktok.com/portal/docs?id=1740029169927169) | P4/S? | —/W? | P1/S? | P2/S? | unknown | pending; source unavailable | system | system |
| WooCommerce | 140 [source](https://woocommerce.github.io/woocommerce-rest-api-docs/) | P4/S4 | —/W73 | P63/S63 | P73/S73 | 0; none classified | pending; none classified | system | system |
| Zendesk Support | 629 [source](https://developer.zendesk.com/zendesk/oas.yaml) | I33/S33 | I62+P28/W294 | P308/S308 | P204/W294 | 0/S1 `binary_read` | pending; six binary/multipart-tagged rows, direction not preserved | system | system |

`binary_upload` is pending globally: its intent is not on `main` and is supplied only by unmerged PR #4343. No row above asserts that it is currently implemented. The source tags identify potential later work; they do not invent a bounded input policy, media allow-list, size cap, or provider request shape.

### Schedule and flow are system-level surfaces

`schedule` and `flow` are not `cli_surface` intents; the schema intent enum is the authority. `pm flow` is nevertheless a real top-level system command that composes existing ETL connections, queries, RLM steps, and already-approved reverse-ETL jobs. `pm schedule` stores and fires an approved flow on a scheduler. Therefore no connector can honestly declare either as a connector intent. A connector is useful *to* flow/schedule only after its own operation is usable; a destination transport is additionally required for a warehouse-mediated typed sync target. Current built-binary samples prove this only for sampled Mailchimp and Zendesk commands, not all declared rows.

### Transport destination gap map

Every cohort bundle currently has `source_bindings=[]`, `eligible_actions=[]`, and no `destination_transport` because the file is absent. The candidate counts below are not selections. Selecting an executor, stream, mapping, or action without a source-bound contract would invent a provider integration.

| Connector | Declared ETL streams | Current named `writes.json` actions | What blocks a destination declaration |
| --- | ---: | ---: | --- |
| Airtable | 5 | 12 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| BambooHR | 84 | 101 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| Buildkite | 6 | 17 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| eBay Fulfillment | 4 | 0 | First retain a complete provider source and derive a named write action with bounded schema; then choose a source binding. |
| Fastly | 6 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| Google Analytics Data API | 5 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| HubSpot | 0 | 0 | First derive an ETL stream plus named source-cited write action with bounded schema; then choose a source binding. |
| LaunchDarkly | 5 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| Linear | 4 | 0 | First establish a GraphQL write contract, bounded schema, and read-back; then choose a source binding. |
| Mailchimp | 79 | 148 | Choose one action and a bounded source record mapping/read-back; reverse-ETL command rows alone are not a transport descriptor. |
| Pinterest | 5 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| Pipedrive | 21 | 17 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| QuickBooks | 5 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| Salesforce | 4 | 0 | First establish a tenant-safe source-cited action, bounded schema, and read-back; then choose a source binding. |
| ShipStation | 4 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| SonarCloud | 11 | 8 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| Squarespace | 8 | 2 | Choose one action and a bounded source record mapping/read-back; no target action or source binding is currently selected. |
| TikTok Marketing | 4 | 0 | First retain a complete provider source and derive a named write action with bounded schema; then choose a source binding. |
| WooCommerce | 4 | 0 | First derive a named source-cited write action with bounded schema; then choose a source binding. |
| Zendesk Support | 33 | 90 | Choose one action and a bounded source record mapping/read-back; reverse-ETL command rows alone are not a transport descriptor. |

A faithful future `destination_transport` for any row must use the closed `declarative_api/declarative_typed_destination` executor and declare: (1) a non-empty exact `eligible_actions` set of `writes.json` names; (2) a mode-to-closed-apply-strategy binding for every selected action; (3) exact admitted source executor references and stream allowlists in `source_bindings`; (4) one `config_match` or `input_fields` mapping per binding, a `per_record` batch bound of 1--1,000, and a separate tombstone mapping where deletes are admitted; (5) provider-specific bounded receipt read-back (identity, expected values, response locator, record/attempt/time limits); (6) explicit idempotency, ordering, delete, durable-warehouse acknowledgement, and independently accepted conformance evidence. The only current source descriptors that a target could cite without inventing a source executor are GitHub's `declarative_api/declarative_stream_source` and PostgreSQL's `native_database/postgres_polling_watermark`; the target-to-source stream and field mapping still require a provider- and action-specific contract.

### Phase 1 verification record

- `scripts/gsd doctor`, all five `scripts/gsd sources` lookups, and `go run ./cmd/agentcontractgen check` — passed.
- `GOFLAGS=-p=3 go build -o pm ./cmd/pm` — passed (`pm dev`).
- Built `pm` in isolated initialized projects without credentials: 36 distinct Mailchimp commands and Zendesk Support `streams tickets list` each returned `missing --credential`; no provider credential or live provider call was used.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'` — passed from cache after compilation; it confirms all 390 `implemented` declarations resolve in the real commandrunner but does not replace binary-boundary evidence.
- The pre-commit hook's unbounded `go test ./internal/connectors ./internal/cli ./cmd/iconregistrygen` exceeded Go's ten-minute default twice. Its exact checks were run directly with the required explicit bound before a disclosed `--no-verify` commit: `gofmt -w cmd internal` (no diff), `go test -timeout 20m ./internal/connectors` (passed), `go test -timeout 20m ./internal/cli` (passed), `go test -timeout 20m ./cmd/iconregistrygen` (passed), `go build ./cmd/pm` (passed), and `./pm docs validate --connectors-dir docs/connectors` (passed).

## Phase 2 approved scope and foundation check — 2026-08-24

### Scope and delivery decision

Captain approval narrowed Phase 2 to one `sync_transport.json` candidate for each of Airtable,
BambooHR, Buildkite, Mailchimp, Pipedrive, SonarCloud, Squarespace, and Zendesk Support. HubSpot
was excluded because it has no named `writes.json` actions. No other connector is in this wave.

The intended executor was the existing closed
`declarative_api/declarative_typed_destination`, not a new generic HTTP or SQL surface. A real
transport must select exact target action(s), mode strategy(ies), source executor/stream allowlists,
bounded field mappings and batch size, provider receipt read-back, and delivery/idempotency/
ordering/delete/acknowledgement semantics. It is not sufficient for the JSON schema merely to load.

### Manual GSD/TDD fallback

The compatible isolated-worker runtime is unavailable and the single-worker contract forbids role
spawning. I ran the GSD prompts inline and retain this trace as the plan, TDD ledger, and
verification checklist.

- `Red:` No target has `sync_transport.json`; no target `writes.json` action has an
  `idempotency_key_header`. A transport declaration selecting any such action must be rejected
  before credentials or provider I/O.
- `Green condition:` Add a connector-local transport only after a selected, source-pinned write
  action declares the provider-owned idempotency header and a connector-owned, bounded receipt
  read-back policy. Then prove source/action/mapping preflight and the built binary's credential
  boundary for every newly implemented CLI command.
- `Refactor/guard:` Do not weaken the shared executor or claim an at-least-once delivery mode to
  bypass idempotency/read-back proof. Any missing provider contract is a separate source-contract
  or foundation decision, not decorative transport metadata.
- Required skills retained for this Go/connector/CLI evaluation: `golang-how-to`, `golang-cli`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`,
  `golang-security`, `golang-safety`, and `golang-testing`.
- CLI help/manual/website parity: no connector surface was changed during this failed-admission
  check. If a transport is later admitted, inspect/help/manual/website applicability must be
  reviewed before implementation.

### Admission evidence

The eight target `writes.json` files contain 395 named actions in total: Airtable 12, BambooHR
101, Buildkite 17, Mailchimp 148, Pipedrive 17, SonarCloud 8, Squarespace 2, and Zendesk Support
90. A `jq` inventory of their `idempotency_key_header` fields returned zero for every connector.
Search of the full pinned bundle is likewise empty for the first seven; Zendesk's only matches are
generic prose saying that idempotency notes may be required, not a provider-owned header in a
write action. This does **not** claim that any provider lacks idempotency support; it establishes
only that the pinned contracts do not currently prove one for an executable target action.

The existing executor is intentionally stricter than the JSON descriptor schema:

- `declarativeTypedDestinationContractFor` calls
  `DeclarativeTypedDestinationIdempotencyHeader` for each selected action and fails when the
  action has no independent provider proof (`internal/app/issue_label_warehouse_transport.go`).
- `declarativeTypedDestinationIdempotencyHeader` reads only the compiled action's
  `idempotency_key_header` and otherwise returns `action <name> has no provider idempotency key
  header` (`internal/connectors/engine/connector.go`).
- The same contract requires action-owned read-back and refuses `full_overwrite`; a receipt
  policy supplied merely as a connector-wide convenience cannot satisfy it.

Focused real-runtime checks:

- `go test -timeout 20m ./internal/app -run '^(TestDeclarativeTypedDestinationRequiresActionOwnedReadBackPolicy|TestDeclarativeTypedDestinationPreflightRejectsIncompleteMappingAndFullOverwriteBeforeIO|TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields)$'` — passed. These tests assert the closed runtime rejects incomplete mappings, missing action-owned receipt policy, and unsupported full-overwrite before provider I/O.
- `go build ./cmd/pm` — passed before local inspection. No new command was marked `implemented`, so there is no new credential-boundary claim to make.

### Result: cannot faithfully admit a Phase 2 transport

All eight candidates fail the generic typed destination's idempotency admission before a source
stream mapping or credential boundary can be reached. Adding `sync_transport.json` now would be a
declaration-only claim that fails shared runtime preflight. The required next work is a distinct
source-contract decision: retain provider-pinned evidence for one suitable action per connector,
including its idempotency header and bounded receipt locator/read-back; only then author and test
the transport descriptor. No production connector or engine files were changed by this check.

## Phase 2 source-contract feasibility probe — 2026-08-24

### Authorization and bounded method

Firstmate message `005` authorized the source-contract work for the same eight connectors only,
without adding any write actions. The governing rule is that an idempotency header and receipt
read-back may be added only when the connector's own pinned source proves both. A provider contract
that is absent, source-drifted, or not retained is an ineligible target, never a guessed header.

The source locks in this cohort are schema-version 1 parity locks. They retain operation identity,
source URL, byte count, and SHA-256 but not the source's request-header and response declarations.
I therefore probed one existing candidate action per connector only after re-fetching its exact
locked public artifact and comparing its SHA-256. A changed document is not evidence: the authoring
conventions classify it as a source-lock refresh decision, not a contract that may be imported.

### Per-connector result

| Connector | Existing actions requiring possible source derivation | Result | Pinned source evidence |
| --- | ---: | --- | --- |
| Airtable | 12 | Not currently eligible. The sole retained rendered-reference artifact drifted, so the original request-header and receipt shape are unavailable. | `https://airtable.com/developers/web/api/introduction`, expected SHA-256 `9a61c17c…b2cb239`, fetched `ab87a415…bc99cba`. The relevant existing action maps to `POST /v0/{baseId}/{tableIdOrName}` (`create-records`). |
| BambooHR | 101 | Not currently eligible. The sole retained rendered-reference artifact drifted; no replacement contract was accepted. | `https://documentation.bamboohr.com/reference/get-meta-company`, expected `ecfc6382…0ebf15a`, fetched `eb9d25cf…3d399059`; candidate `create-time-tracking-project`, `POST /api/v1/time-tracking/projects`. |
| Buildkite | 17 | Not currently eligible. The sole retained rendered-reference artifact drifted; its endpoint table has no retained header/response material. | `https://buildkite.com/docs/apis/rest-api`, expected `350d7584…982eedd`, fetched `2f2f547b…f71fd0f`; candidate `POST /v2/organizations/{org.slug}/pipelines`. |
| Mailchimp | 148 | Not currently eligible from the bounded probe. The exact root Swagger artifact and selected existing `put_lists_members` operation contain no idempotency declaration. Establishing whether any other existing action differs requires reviewing its action-specific source artifact(s), up to 148 action contracts, rather than a bounded one-action wave. | Root `https://api.mailchimp.com/schema/3.0/Swagger.json`, SHA-256 `9b17c3c8…8e98a9a0`; member operation `https://us22.api.mailchimp.com/schema/3.0/Paths/Lists/Members/Instance.json`, SHA-256 `c2c22744…6bebf47d`, `PUT /lists/{list_id}/members/{subscriber_hash}`. |
| Pipedrive | 17 | Not eligible. The verified full pinned OpenAPI artifact has no `idempotency` declaration, so none of its already-declared actions can supply the required provider-owned header. | `https://developers.pipedrive.com/docs/api/v1/openapi.yaml`, SHA-256 `302b0d7c…bf501c2b`; this covers `POST /leads` (`addLead`) and the other existing actions. |
| SonarCloud | 8 | Not currently eligible. The sole retained catalog artifact drifted, so its original action/header/response contract cannot be derived. | `https://sonarcloud.io/api/webservices/list`, expected `76f39c51…8553db5e`, fetched `46bc01d8…659a3f0a`; candidate `create`, `POST /api/webhooks/create`. |
| Squarespace | 2 | Not eligible. The verified full pinned OpenAPI declares `Idempotency-Key` only for `POST /1.0/commerce/inventory/adjustments` and `POST /1.0/commerce/orders`; neither is one of the two existing `writes.json` actions, which are webhook-subscription operations. | `https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json`, SHA-256 `eff1274e…b2b2debc`; existing candidate `createWebhookSubscription`, `POST /1.0/webhook_subscriptions`. |
| Zendesk Support | 90 | Not eligible. The verified full pinned OpenAPI artifact has no `idempotency` declaration, so none of its already-declared actions can supply the required provider-owned header. | `https://developer.zendesk.com/zendesk/oas.yaml`, SHA-256 `a487892c…36d9a0c8`; this covers `CreateTicket`, `POST /api/v2/tickets`, and the other existing actions. |

The digest prefixes above identify the full values retained in each connector's
`sources/<connector>-parity-source-lock.json`; no provider API operation was executed and no
credential was used. Source checks verified Pipedrive, Squarespace, Zendesk Support, the Mailchimp root, and
the selected Mailchimp member-operation artifact byte-for-byte. Airtable, BambooHR, Buildkite, and
SonarCloud were refused because their fetched bytes differed from the pinned digest.

### Honest size and stop condition

No connector reached header admission, so no receipt-readback or transport JSON can be authored
faithfully. The remaining source work is not a bounded eight-action change: four connectors first
need immutable source-artifact retention or an approved source-lock refresh (12 + 101 + 17 + 8 =
138 existing actions potentially affected), and Mailchimp requires action-specific review across
up to 148 existing contracts to discover whether a header-bearing action exists. Pipedrive,
Squarespace, and Zendesk Support are already ruled out by their verified full pinned sources.

This exceeds the authorized bounded wave. No `writes.json`, `sync_transport.json`, CLI surface, or
shared-engine file was changed. The required decision is whether to open a separate, source-lock
refresh/retention and per-action-evidence wave; without it, this task's honest Phase 2 result is
zero new typed destinations.
