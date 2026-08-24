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
| Binary upload is not accidentally claimed for this batch | live | Current `main` has the #4343 binary-upload intent and one GitHub command; none of these twenty connectors has an implemented binary-upload command. |

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
- The formerly reported 7,301-operation source-lock aggregate is historical audit data, not a current provider-contract total: five of its counted URLs were later proved to be landing pages. No cohort-wide source-operation total is asserted until the replacement sources are committed as new lock identities and their projections are re-derived. eBay Fulfillment and TikTok Marketing have no retrievable public source; their totals are unknown, not zero.
- Eight, rather than seven, bundles have `writes.json`: Airtable, BambooHR, Buildkite, Mailchimp, Pipedrive, SonarCloud, Squarespace, and Zendesk Support. All twenty lack `sync_transport.json`.
- `TestEveryImplementedCommandPassesRuntimePreflight` succeeds for all 390 declared rows. It is structural evidence only. Built-binary credential-boundary probes independently confirmed 36 distinct Mailchimp commands and Zendesk Support's `streams tickets list`; the remaining 353 declared implemented rows are not counted as Phase 1 binary-certified usable.

### Source-URL correction — 2026-08-24

After approval of the initial map, Firstmate remeasured every locked URL and established that five
of this cohort's recorded URLs resolve to **landing pages, not provider specifications**:
**Google Analytics Data API, Linear, Mailchimp, QuickBooks, and Salesforce.** The counts and lane
crosswalks marked `†` below are therefore historical observations from the checked-in lock, **not
admissible provider-contract evidence for Phase 2**. Firstmate's two-fetch research has now
supplied actual replacements, but this source-only correction deliberately does not import, pin,
or project them:

- Google Analytics Data API — `https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta`
- Linear — `https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql`
- Mailchimp — `https://mailchimp.com/developer/marketing/api/`
- QuickBooks — `https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities`
- Salesforce — `https://developer.salesforce.com/docs/get_document_content/api_rest/resources_list.htm/en-us/262.0`

The same research confirms that **eBay Fulfillment** and **TikTok Marketing** have no retrievable
public source (`‡` below): eBay's direct OAS, release-notes, and rendered-reference requests were
403 responses; TikTok's legacy and current provider-documentation requests failed TLS in both
curl and Chrome. No operation, `not_applicable` claim, source transition, or executable surface
will be re-derived until a provider source can be retrieved. In particular, retaining a
byte-identical historical artifact does not make its landing-page URL a usable specification
source.

Notation below: `I` means declared `implemented`, `P` means declared `partial`, `S` means the source disposition crosswalk count, and `W` means a source-backed provider write candidate. `S` can include explicit local execution bindings; the source column remains the provider-operation total. `none classified` never claims a provider lacks a capability: the pinned lock simply does not establish a bounded binary contract.

| Connector | Pinned source total | ETL | Reverse ETL | Direct read | Direct write | Binary download | Binary upload | Schedule | Flow |
| --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| Airtable | 103 [source](https://airtable.com/developers/web/api/introduction) | P5/S5 | —/W74 | P26/S26 | P74/S74 | 0; none classified | pending; none classified | system | system |
| BambooHR | 319 [source](https://documentation.bamboohr.com/reference/get-meta-company) | P84/S84 | —/W190 | P71/S71 | P190/S190 | 0; none classified | pending; 19 binary/multipart-tagged rows, direction not preserved | system | system |
| Buildkite | 129 [source](https://buildkite.com/docs/apis/rest-api) | P6/S6 | —/W76 | P50/S50 | P76/S76 | 0; none classified | pending; none classified | system | system |
| eBay Fulfillment ‡ | unknown; no retrievable public source | P4/S? | —/W? | P4/S? | P3/S? | unknown | pending; source unavailable | system | system |
| Fastly | 732 [source](https://www.fastly.com/documentation/downloads/fastly.collection.json) | P6/S6 | —/W389 | P337/S337 | P389/S389 | 0; none classified | pending; none classified | system | system |
| Google Analytics Data API † | unknown; [replacement Discovery v1beta](https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta) | P5/S? | —/W? | P7/S? | P16/W? | unknown | pending; source projection not yet derived | system | system |
| HubSpot | 3,118 [source](https://codeload.github.com/HubSpot/HubSpot-public-api-spec-collection/tar.gz/2bebde2dca45eaa1792931089c4e441c8e377594) | 0/S0 | —/W1,901 | P1,240/S1,217 | P1,901/S1,901 | 0/S32 `binary_read` | pending; 229 binary/multipart-tagged rows, direction not preserved | system | system |
| LaunchDarkly | 397 [source](https://app.launchdarkly.com/api/v2/openapi.json) | P5/S5 | —/W205 | P189/S189 | P205/S205 | 0; none classified | pending; none classified | system | system |
| Linear † | unknown; [replacement GraphQL SDL](https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql) | P4/S? | —/W? | P166/S? | P373/W? | unknown | pending; source projection not yet derived | system | system |
| Mailchimp † | unknown; [replacement Marketing API reference](https://mailchimp.com/developer/marketing/api/) | I79/S? | I148/W? | I68+P12/S? | P16/W? | unknown | pending; source projection not yet derived | system | system |
| Pinterest | 279 [source](https://developers.pinterest.com/docs/api/v5/introduction/) | P5/S5 | —/W135 | P144/S144 | P135/S135 | 0; none classified | pending; none classified | system | system |
| Pipedrive | 213 [source](https://developers.pipedrive.com/docs/api/v1/openapi.yaml) | P21/S21 | —/W99 | P98/S98 | P99/S99 | 0; one explicit raw-binary download | pending; 12 binary/multipart-tagged rows, direction not preserved | system | system |
| QuickBooks † | unknown; [replacement Accounting API reference](https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities) | P5/S? | —/W? | P84/S? | P45/W? | unknown | pending; source projection not yet derived | system | system |
| Salesforce † | unknown; [replacement REST resource reference](https://developer.salesforce.com/docs/get_document_content/api_rest/resources_list.htm/en-us/262.0) | P4/S? | —/W? | P3/S? | P3/W? | unknown | pending; source total remains tenant-dynamic | system | system |
| ShipStation | 47 [source](https://docs.shipstation.com/_bundle/apis/@shipstation-v1/openapi.json?download) | P4/S4 | —/W25 | P18/S18 | P25/S25 | 0; none classified | pending; none classified | system | system |
| SonarCloud | 156 [source](https://sonarcloud.io/api/webservices/list) | P11/S11 | —/W87 | P59/S59 | P87/S87 | 0; none classified | pending; none classified | system | system |
| Squarespace | 53 [source](https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json) | P8/S8 | —/W30 | P15/S15 | P30/S30 | 0; none classified | pending; one multipart image-upload row | system | system |
| TikTok Marketing ‡ | unknown; no retrievable public source | P4/S? | —/W? | P1/S? | P2/S? | unknown | pending; source unavailable | system | system |
| WooCommerce | 140 [source](https://woocommerce.github.io/woocommerce-rest-api-docs/) | P4/S4 | —/W73 | P63/S63 | P73/S73 | 0; none classified | pending; none classified | system | system |
| Zendesk Support | 629 [source](https://developer.zendesk.com/zendesk/oas.yaml) | I33/S33 | I62+P28/W294 | P308/S308 | P204/W294 | 0/S1 `binary_read` | pending; six binary/multipart-tagged rows, direction not preserved | system | system |

Historical correction: this initial map predated the #4343 merge. `binary_upload` is now an intent
on `main`, with one implemented GitHub command, but none of these twenty connectors implements it.
The table's `pending` cells now mean only that the pinned provider source may identify future
connector-local work; they do not claim a missing global engine feature, a bounded input policy,
media allow-list, size cap, or provider request shape.

`†` Firstmate's two-fetch source research (`data/cli-top100-source-research-r1/report.md`, rows
31, 34, 36, 39, and 43) proved that the old URL was not the provider specification and supplied
the linked replacement. The local declaration counts remain measured facts; all old provider
operation and `S`/`W` values are retired from this map until a new source lock and projection
exist. `‡` The same report's rows 42 and 49 establish that no retrievable public provider source
exists for TikTok Marketing or eBay Fulfillment, respectively; their source-dependent values
remain unknown rather than being inferred from an error or shell page.

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
| Airtable | 12 | Not currently eligible. The sole retained rendered-reference artifact drifted, so the original request-header and receipt shape are unavailable. | `https://airtable.com/developers/web/api/introduction`, expected SHA-256 `9a61c17c…b2cb239`, fetched `70a3f709…c38322a5`. The relevant existing action maps to `POST /v0/{baseId}/{tableIdOrName}` (`create-records`). |
| BambooHR | 101 | Not currently eligible. The sole retained rendered-reference artifact drifted; no replacement contract was accepted. | `https://documentation.bamboohr.com/reference/get-meta-company`, expected `ecfc6382…0ebf15a`, fetched `a04e6544…f1177890`; candidate `create-time-tracking-project`, `POST /api/v1/time-tracking/projects`. |
| Buildkite | 17 | Not currently eligible. The sole retained rendered-reference artifact drifted; its endpoint table has no retained header/response material. | `https://buildkite.com/docs/apis/rest-api`, expected `350d7584…982eedd`, fetched `c969b9e0…e07262d2`; candidate `POST /v2/organizations/{org.slug}/pipelines`. |
| Mailchimp | 148 | Not currently eligible from the bounded probe. The exact root Swagger artifact and selected existing `put_lists_members` operation contain no idempotency declaration. Establishing whether any other existing action differs requires reviewing its action-specific source artifact(s), up to 148 action contracts, rather than a bounded one-action wave. | Root `https://api.mailchimp.com/schema/3.0/Swagger.json`, SHA-256 `9b17c3c8…8e98a9a0`; member operation `https://us22.api.mailchimp.com/schema/3.0/Paths/Lists/Members/Instance.json`, SHA-256 `c2c22744…6bebf47d`, `PUT /lists/{list_id}/members/{subscriber_hash}`. |
| Pipedrive | 17 | Not eligible. The verified full pinned OpenAPI artifact has no `idempotency` declaration, so none of its already-declared actions can supply the required provider-owned header. | `https://developers.pipedrive.com/docs/api/v1/openapi.yaml`, SHA-256 `302b0d7c…bf501c2b`; this covers `POST /leads` (`addLead`) and the other existing actions. |
| SonarCloud | 8 | Not currently eligible. The sole retained catalog artifact drifted, so its original action/header/response contract cannot be derived. | `https://sonarcloud.io/api/webservices/list`, expected `76f39c51…8553db5e`, fetched `f160f7f6…c71ca4080`; candidate `create`, `POST /api/webhooks/create`. |
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

## Write-action fidelity correction — 2026-08-24

Firstmate requested a correctness audit of Pipedrive, Squarespace, and Zendesk Support because the
prior wording could imply that the verified sources refute their shipped write actions. That
inference was wrong: the sources refute *typed-destination idempotency admission*, not the action
method/path contracts. The per-action audit is in
[`cli-batch45-make-usable-r1-write-action-audit.md`](cli-batch45-make-usable-r1-write-action-audit.md).

- Pipedrive: 17/17 declared action method/path identities match the verified full OpenAPI source.
- Squarespace: 2/2 match after the declaration's `base_url` `/1.0` prefix is applied.
- Zendesk Support: 90/90 match after path-variable labels are compared by position; e.g. a
  declaration's `{id}` is the source operation's typed resource identifier, not a different wire
  segment.

No source method/path contradiction was observed. A non-idempotent action may still make one
provider-correct request; the engine disables its retries. Neither that fact nor source identity
certifies every declared CLI row as binary-reachable, so the Phase 1 credential-boundary evidence
is unchanged.

Mailchimp has exactly 148 existing `writes.json` actions. A full source-contract review is 148
action-level reviews, not a rough estimate: for each action, retain and digest-verify its own
provider artifact; derive a provider-owned idempotency header, if one is actually declared; derive
the write response's bounded receipt locator; identify a separately declared, bounded provider
read-back operation and its input/response mapping; and then prove the candidate's exact source
stream field mapping and delivery limits. The verified root Swagger artifact and the selected
`put_lists_members` artifact declare no idempotency header; they do not prove every other
action-specific artifact has the same absence.

## Captain priority: source locks before surfaces — 2026-08-24

Captain direction supersedes further Phase 2 transport work. The primary deliverable is now a
twenty-connector source-lock inventory and remediation plan: record lock presence, old and new
bytes/digests for every re-pin, the exact retained source/spec URL, and whether Context7
documentation evidence was needed because a provider source is missing, broken, or prose-only.
No transport, write-action, or parallel source-refresh implementation will be authored here. PR
#4348 owns the source-retention foundation; re-pinning will consume that foundation after it lands.

- `Red:` Existing parity locks prove only operation inventories for several connectors; some public
  artifacts now drift and eBay/TikTok/Salesforce do not yet have a valid retained provider source.
- `Green condition:` Every connector has an exact, provider-owned source URL and a retained,
  hash-verifiable artifact or an explicit unavailable/source-evidence record. Context7 may supply
  cited documentation discovery only; it cannot substitute for a retained provider artifact.
- `Guard:` A changed artifact is reported as a re-pin with both old and new bytes/digests, never as
  if it remained the old pin. No guessed documentation URL, operation, header, or transport is
  accepted.

### Source-lock inventory — 20 connectors

This is a **source-lock report**, not a source refresh. Every connector has a schema-version-1
`sources/<connector>-parity-source-lock.json`, but that format retains only source metadata and
normalized local inventory; it does not retain the provider bytes. Thus `match` below means a
single current re-fetch produced the lock's recorded byte count and SHA-256. It does *not* make the
current tree hermetic. PR #4348 is the designated retention foundation; no lock file is changed in
this branch.

All current fetches below produced the byte count and SHA-256 from the same response body. A
`re-pin candidate` is deliberately not a re-pin: its artifact still needs the PR #4348 retention
path, content validation, and an explicit old-to-new provenance change. `root-only` states the
checked extent where a lock has multiple provider artifacts.

| Connector | Lock / exact provider source | Current verification (old → current bytes, SHA-256) | Context7 and disposition |
| --- | --- | --- | --- |
| Airtable | Present; `https://airtable.com/developers/web/api/introduction` | **drift / re-pin candidate:** `623469`, `9a61c17ca297d70ba6ec186a7acb03a00d15915b3007bc11db6263cf0b2cb239` → `623785`, `70a3f7090044871d48e050620c6c0ee74ffc91013e4cae40373b4999c38322a5` | Used: confirms the official Web API reference. Rendered provider documentation; retain a browser capture only through #4348. |
| BambooHR | Present; `https://documentation.bamboohr.com/reference/get-meta-company` | **drift / re-pin candidate:** `1562871`, `ecfc63823d7f08942bec89f7175ac6fedc582b07177346fdc8e4d03400ebf15a` → `1627565`, `a04e654442c3980b7f3172b7160404723a1ae096ca78aca1cead0721f1177890` | Used: confirms the official REST reference. Rendered-reference retention is required. |
| Buildkite | Present; `https://buildkite.com/docs/apis/rest-api` | **drift / re-pin candidate:** `440978`, `350d758449efdfcf9e0522cf10731c824a8cde8570869187a2ea1bb95982eedd` → `439436`, `c969b9e057a29c90321c0615264e67f2c08c767c94fc6b5770ccda77e07262d2` | Used: confirms the official REST reference. Rendered-reference retention is required. |
| eBay Fulfillment | Present, but **zero retained artifacts**; old stale route `https://developer.ebay.com/develop/api/fulfillment-api/release-notes`; actual current documentation `https://developer.ebay.com/develop/api/sell/fulfillment_api` | **no comparable pin:** no old bytes/digest. The stale URL returned a 1,830-byte error page, SHA-256 `cc4e52c700e9f7e25c5d676263a91e4e008b9760129b4f22c9a90f803519b51c`; it is not source evidence. | Used: official eBay documentation confirms the Fulfillment API's live endpoint catalog and that a downloadable OAS existed, but no current first-party download URL was discovered or retained. Capture the current official reference through #4348; do not retain the error page. |
| Fastly | Present; `https://www.fastly.com/documentation/downloads/fastly.collection.json` | **match:** `2341028`, `c6ae5b0fd118fe2d87e7d0ef431f67cda703d1487d27b5f02725d82219386677` → identical | Not needed: machine-readable provider collection; retain its exact response under #4348. |
| Google Analytics Data API | Present; `https://analyticsdata.googleapis.com/$discovery/rest?version=v1` (also locked: `v1alpha`, `v1beta`) | **drift (v1 lock):** `14971`, `92487efec56020b9ef2f9644b55a352f9ae9ca34eba41612f90a82d54266b499` → `14971`, `da7d18b86488d5d8fcceda6f13774284e88e7c70da2d26dbffcec00a885c3e13` | Not needed: provider Discovery documents are machine-readable. A single drift means the three-artifact lock must be revalidated as a set before any re-pin. |
| HubSpot | Present; `https://codeload.github.com/HubSpot/HubSpot-public-api-spec-collection/tar.gz/2bebde2dca45eaa1792931089c4e441c8e377594` | **match:** `4132728`, `7bfcdb27e8d7e52341e90284f670768020f62aeebf43f5cd339263fa5d801619` → identical | Not needed: immutable-commit provider-owned spec collection; retain the exact tarball under #4348. |
| LaunchDarkly | Present; `https://app.launchdarkly.com/api/v2/openapi.json` | **drift / re-pin candidate:** `2953979`, `41fc8c76b779790f405bc0f1f500ab54a6a3695bb14a9916c537c180e5469419` → `2936808`, `5712a80c347abfc6a731296e487eee0e9d282ef910ba10c3ff16949daa8d9b3e` | Not needed: machine-readable provider OAS; validate and retain the candidate through #4348. |
| Linear | Present; `https://studio.apollographql.com/public/Linear-API/variant/current/schema/reference` | **drift / browser-capture candidate:** `892766`, `ff3b49156874dd6d01d12541828bb18210cb7e617de577097505895eb3312c7e` → rendered DOM `889697`, `e7585a99cc8dc0fb8a745690d473d3d0ee561ebb06b7b1b08d5f3228445f9dfd`. A direct HTTP fetch returns only a 7,460-byte application shell and is not comparable. | Used: confirms Linear's official GraphQL endpoint (`https://api.linear.app/graphql`) and public schema reference. The public schema page renders without login; retain a browser capture, not the shell, through #4348. |
| Mailchimp | Present; root OAS `https://api.mailchimp.com/schema/3.0/Swagger.json` plus 181 operation artifacts | **root-only match:** `73813`, `9b17c3c80104e6ca41ff3b65286640bf4d698793e832779901167a9a8e98a9a0` → identical. The other 181 locked artifacts have not been claimed matched. | Not needed: machine-readable root and operation artifacts. #4348 must retain and verify all 182 as a set before certifying the lock. |
| Pinterest | Present; `https://developers.pinterest.com/docs/api/v5/introduction/` | **drift / re-pin candidate:** `651974`, `2fd707bc8df87440903b46711e836370da639f485078a600019760f3e29a6d63` → `657429`, `1a5d73b752b5839c0853a2f481c85bf99200c6175cffd177e24a0f0ae231dece` | Used: confirms the official v5 developer reference. Retain a validated browser/rendered capture through #4348. |
| Pipedrive | Present; `https://developers.pipedrive.com/docs/api/v1/openapi.yaml` | **match:** `1782400`, `302b0d7c2c1a6cb96a2d299717c6be0c2cf3eac6dfd884ea8352962ebf501c2b` → identical | Not needed: machine-readable provider OAS; retain exact bytes under #4348. |
| QuickBooks | Present; stale artifact `https://static.developer.intuit.com/JSONObjects/EntityJsonObject_v1.json`; actual provider documentation uses the QuickBooks Online API Explorer at `https://developer.intuit.com/app/developer/qbo/docs/develop` | **broken, not a re-pin candidate:** `387314`, `24c6accfab8236fdba4f03bff33214dab5b891e3a0c82b243cf6ca4f297fdb7a` → `33372`, `228748c6533cdbaeeb9e18cde987ea68e7ce1de6f66cdd4e6fd655da71049583`; the current response is not the former entity contract. | Used: official developer material identifies the API Explorer as the current schema surface; no downloadable first-party accounting OAS/JSON replacement was discovered. Record a source-evidence transition, not a content re-pin. |
| Salesforce | Present, browser-rendered source metadata but **zero retained artifacts**; `https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_list.htm` | **not comparable / dynamic source:** browser lock `76875`, `609709e1726e5ff221f8b86d9f7408c5acb0b00e72e99e3ecc4d6635ecadf474`; current direct response `17891`, `b4b00990d9e4fc091eb9365cd7a6e52391eaaaa0f1bb33067150c14763293389`. These different acquisition modes are not a valid re-pin pair. | Used: Salesforce's REST resources are tenant/version dependent under `https://<my-domain>.my.salesforce.com/services/data/vXX.0/`; no credential-free global operation total or retainable global REST OAS can be asserted. Its sObject OAS generation is per-org, so it is out of scope without credentials. |
| ShipStation | Present; `https://docs.shipstation.com/_bundle/apis/@shipstation-v1/openapi.json?download` | **match:** `186490`, `c71c20d26559b1d0dad5c3718889fc8d4063cfad3abeaf5f63cc1651bef32307` → identical | Not needed: machine-readable provider OAS; retain exact bytes under #4348. |
| SonarCloud | Present; `https://sonarcloud.io/api/webservices/list` | **drift / re-pin candidate:** `209845`, `76f39c511c7ab51254d2d2032baf9b1ff062d37a6b2eac766c6dba2d8553db5e` → `211930`, `f160f7f6764d6750544e2dde53acba1a7ffaac2d8912c773b7c92b2c71ca4080` | Not needed: provider web-service catalog is fetchable. Validate and retain the candidate through #4348. |
| Squarespace | Present; `https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json` | **match:** `345839`, `eff1274e6e87cfa998a5125c2ebf53ee459202d108598dacf6507b32b2b2debc` → identical | Not needed: machine-readable provider OAS; retain exact bytes under #4348. |
| TikTok Marketing | Present, but **zero retained artifacts**; stale route `https://business-api.tiktok.com/portal/docs?id=1740029169927169`; current official guide `https://business-api.tiktok.com/gateway/docs/index?doc_id=1735713875563521&identify_key=c0138ffadd90a955c1f0670a56fe348d1d40680b3c89461e09f78ed26785164b&language=ENGLISH` | **no comparable pin:** no old bytes/digest. The stale route still fails TLS in a direct client; no response body was accepted. | Used: the official v1.3 guide publishes the base URL `https://business-api.tiktok.com/open_api/v1.3` and a provider Postman collection, but no first-party OAS download was found. Browser capture/source discovery through #4348 is required; do not use a third-party OAS. |
| WooCommerce | Present; `https://woocommerce.github.io/woocommerce-rest-api-docs/` | **match:** `4400931`, `a02e504ae56786d94f6a859e55b5c7e3229749ea3d43147424cfd987b2ec550e` → identical | Used: confirms the official REST reference. Retain the exact rendered artifact through #4348. |
| Zendesk Support | Present; `https://developer.zendesk.com/zendesk/oas.yaml` | **match:** `1757202`, `a487892c8e1f3feeba96c234148be69fddd50afce17bf30437bcb8de36d9a0c8` → identical | Not needed: machine-readable provider OAS; retain exact bytes under #4348. |

### Counted result and dependency order

- **20/20** connector parity-lock metadata files exist; **0/20** currently retain their provider
  response bytes in the bundle.
- **7 first-artifact matches:** Fastly, HubSpot, Pipedrive, ShipStation, Squarespace, WooCommerce,
  Zendesk Support. Mailchimp is an additional **root-only** match (1 of 182 artifacts).
- **7 valid changed-artifact re-pin candidates:** Airtable, BambooHR, Buildkite, Google Analytics
  Data API, LaunchDarkly, Pinterest, SonarCloud. Each carries the full old/new byte-and-digest pair
  above and must be treated as a re-pin, never as a match.
- **6 source-evidence repairs:** eBay Fulfillment, Linear, QuickBooks, Salesforce, TikTok
  Marketing, plus the remaining 181 Mailchimp operation artifacts. These need source discovery or
  complete artifact verification, not a digest substitution. (Linear has a candidate public browser
  capture, but not a raw GraphQL contract.)

The only safe dependency order is: land #4348's retention mechanism; validate/retrieve the matched
and candidate source artifacts through that mechanism; write explicit provenance updates for each
re-pin; then regenerate/compare the local operation mapping. No command, write action, or transport
may be promoted on the basis of a source URL alone.

## Source-retention compatibility continuation — 2026-08-24

PR #4348 landed at `db2892653`, and Firstmate directed this lane to use its `source-retain`
maintenance command. The first controlled attempt is the TDD red observation:

```text
go run ./cmd/connectorgen source-retain fastly \
  --retrieved-at 2026-08-24T12:05:00Z --license undetermined --terms undetermined
→ resolve connector-owned source lock: .../fastly-operation-source-lock.json: no such file or directory
```

All 20 Batch 4/5 connectors have only their prior `*-parity-source-lock.json`; none has an
`*-operation-source-lock.json`. The landed command correctly refuses to guess a URL, but it has no
reader for the already connector-owned parity artifact identity. That makes retention impossible
even for Fastly, whose current provider collection matched the parity lock byte-for-byte.

### Manual GSD / TDD plan

The canonical GSD command sources were resolved with `scripts/gsd doctor`, `scripts/gsd sources`
for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, and the
corresponding `scripts/gsd prompt … 4349` prompts were reviewed. This task uses the documented
inline/manual fallback: the connector delivery contract forbids role spawning, and this is a small
shared maintenance compatibility hook rather than a connector runtime change. Required skills
loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, and `golang-documentation`; the CLI help/docs parity guidance
was also read.

- **Scope:** `cmd/connectorgen` only plus its tests and the source-retention documentation. No
  connector runtime, engine, CLI surface, or provider contract changes.
- **Red:** add a test that a valid connector-owned v1 parity lock with one exact public artifact is
  retained and recorded without modifying that parity lock. Add a regression that the presence of
  an operation lock is authoritative: a malformed operation lock must fail rather than falling back
  to a parity lock. Add a fixed public-query case so identity queries already present in a parity
  artifact are retained as part of identity, never accepted from command input.
- **Green:** let `source-retain` read the existing operation lock when it exists; otherwise read
  only `<connector>-parity-source-lock.json`, duplicate-safe decode its recorded artifact list,
  validate connector identity, URL, bytes, and SHA-256 through the existing public-URL policy, and
  retain only those predeclared artifacts. The compatibility reader has no operation importer
  semantics: `source-import` remains operation-lock-only and hermetic.
- **Guard:** an absent artifact list, an unavailable diagnostic lock, invalid identity, malformed
  lock, terminal response, redirect, login wall, or byte/digest change is terminal. The fallback
  never invents an `operation-source-lock.json`, operation inventory, source URL, or query. A
  query already present in a strict legacy artifact identity is allowed only if the existing public
  query policy accepts it.
- **CLI parity:** `connectorgen` is a developer tool, not `pm`; its own help and
  `docs/migration/conventions.md` are updated and tested. `pm` help/manual/website changes are
  not applicable.

Only after this compatibility hook is green will this lane invoke `source-retain` against the
twenty existing connector-owned parity locks. Each source mismatch will be classified and recorded
before any lock update; `source-retain` itself still cannot alter a lock.

### TDD ledger — legacy parity retention compatibility

- **Red:** the unmodified `source-retain fastly` command failed before any provider request because
  `fastly-operation-source-lock.json` does not exist. The new legacy-parity behavior tests likewise
  failed with that exact missing-operation-lock error.
- **Green:** `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceRetain'` passes. It proves
  exact legacy bytes are retained without modifying the parity lock, preserves a predeclared fixed
  public query as identity provenance, refuses a malformed operation lock instead of falling back,
  and renders the developer-command help contract.
- **Security review target:** the sole new trust boundary is a checked-in legacy source lock. The
  reader uses the existing public URL/SSRF policy, requires a matching connector name and v1 lock,
  allows only the lock's own non-secret query, and runs the same byte/digest validation before any
  write. It never accepts a runtime URL, artifact path, or credentials.

### Retention execution — 2026-08-24

This execution supersedes the *current verification* classifications in the earlier inventory; the
earlier values remain the historical audit observations. `source-retain` was invoked with
`--retrieved-at 2026-08-24T12:05:15Z --license undetermined --terms undetermined`. It receives no
provider URL or credentials. It fetched only the source URL already recorded by each connector,
rejected redirects/non-200 bodies, validated byte count and SHA-256 before writing, then created
the connector-owned artifact and provenance manifest.

| Connector | Result | Retained/re-pin evidence |
| --- | --- | --- |
| Airtable | Refused; no re-pin. The rendered document is not stable enough to lock honestly: the old `623469` / `9a61c17ca297d70ba6ec186a7acb03a00d15915b3007bc11db6263cf0b2cb239` differs from two 623785-byte reads, `70a3f7090044871d48e050620c6c0ee74ffc91013e4cae40373b4999c38322a5` and `0705bc2fd16cd02296778fc48a66ecfb5d1351fb4bb289c37021b6a7ed2d587a`. | `source-retain` refused the old pin; no artifact or manifest was written. |
| BambooHR | **Re-pinned and retained.** | Old `1562871` / `ecfc63823d7f08942bec89f7175ac6fedc582b07177346fdc8e4d03400ebf15a` → real provider reference `1627565` / `a04e654442c3980b7f3172b7160404723a1ae096ca78aca1cead0721f1177890`; artifact and manifest retained. |
| Buildkite | Refused; no re-pin. The rendered document produced different current identities: old `440978` / `350d758449efdfcf9e0522cf10731c824a8cde8570869187a2ea1bb95982eedd`, then `439436` / `c969b9e057a29c90321c0615264e67f2c08c767c94fc6b5770ccda77e07262d2`, then `439436` / `3c8895da56b4ee032cc7eeffe1dab906fef96f95ad762984bbb733734c310dd3`. | `source-retain` refused the old pin; no artifact or manifest was written. |
| eBay Fulfillment | Unmappable. | The parity lock has no retained artifact identity; `source-retain` correctly reported `parity source lock has no retainable provider artifacts`. The current official API page is recorded above, but no error-page or guessed OAS was pinned. |
| Fastly | **Retained unchanged.** | `c6ae5b0fd118fe2d87e7d0ef431f67cda703d1487d27b5f02725d82219386677`, 2341028 bytes. |
| Google Analytics Data API | Refused; no re-pin. The locked v1 discovery document changed despite its byte count remaining 14971: old `92487efec56020b9ef2f9644b55a352f9ae9ca34eba41612f90a82d54266b499`, then `da7d18b86488d5d8fcceda6f13774284e88e7c70da2d26dbffcec00a885c3e13`, then `893dcd5a0f81bfc260f84bbb396011f04c7756803094e5f4e7a6228f2098680b`. | `source-retain` refused the existing multi-artifact lock before write; no artifact or manifest was written. |
| HubSpot | **Retained unchanged.** | `7bfcdb27e8d7e52341e90284f670768020f62aeebf43f5cd339263fa5d801619`, 4132728 bytes. |
| LaunchDarkly | **Re-pinned and retained.** | Old `2953979` / `41fc8c76b779790f405bc0f1f500ab54a6a3695bb14a9916c537c180e5469419` → real provider OAS `2936808` / `5712a80c347abfc6a731296e487eee0e9d282ef910ba10c3ff16949daa8d9b3e`; artifact and manifest retained. |
| Linear | Refused; no re-pin. | The public browser-rendered schema remains a valid citation, but the maintenance HTTP fetch sees the 7460-byte application shell rather than the old rendered capture. `source-retain` rejected it; no shell artifact was pinned. |
| Mailchimp | Refused as an atomic 182-artifact set. | The root Swagger match is insufficient: at least one of the remaining locked operation artifacts now differs. `source-retain` wrote nothing, so no partial root-only lock is falsely presented as retained. |
| Pinterest | Refused; no re-pin. The rendered reference changed from old `651974` / `2fd707bc8df87440903b46711e836370da639f485078a600019760f3e29a6d63`, through `657429` / `1a5d73b752b5839c0853a2f481c85bf99200c6175cffd177e24a0f0ae231dece`, to `657190` / `1e8062a4c822de21589e334778a0135c3ebd7634be1faabf8dd7f98204bf2d1a`. | `source-retain` refused the old pin; no artifact or manifest was written. |
| Pipedrive | **Retained unchanged.** | `302b0d7c2c1a6cb96a2d299717c6be0c2cf3eac6dfd884ea8352962ebf501c2b`, 1782400 bytes. |
| QuickBooks | **Retained unchanged.** | `source-retain` obtained the original `387314`-byte `24c6accfab8236fdba4f03bff33214dab5b891e3a0c82b243cf6ca4f297fdb7a` provider entity document. An earlier ad-hoc direct-client body of 33372 bytes was not accepted or used; it is not a re-pin. |
| Salesforce | Unmappable. | The browser-rendered provenance has no artifact list, so `source-retain` correctly reported no retainable provider artifacts. Tenant-specific REST surface discovery remains unavailable without credentials. |
| ShipStation | **Retained unchanged.** | `c71c20d26559b1d0dad5c3718889fc8d4063cfad3abeaf5f63cc1651bef32307`, 186490 bytes; the locked `download` query is recorded as identity provenance. |
| SonarCloud | Refused; no re-pin. The provider catalog changed from old `209845` / `76f39c511c7ab51254d2d2032baf9b1ff062d37a6b2eac766c6dba2d8553db5e`, through `211930` / `f160f7f6764d6750544e2dde53acba1a7ffaac2d8912c773b7c92b2c71ca4080`, to `211920` / `701e57f73a1d91c1ceea91301e7de8e043e0d5e46b86e637fc289502d1db13a9`. | `source-retain` refused the old pin; no artifact or manifest was written. |
| Squarespace | **Retained unchanged.** | `eff1274e6e87cfa998a5125c2ebf53ee459202d108598dacf6507b32b2b2debc`, 345839 bytes. |
| TikTok Marketing | Unmappable. | The parity lock has no artifact identity and the stale direct documentation route still fails TLS. `source-retain` wrote nothing; no third-party OAS or failure response was substituted. |
| WooCommerce | **Retained unchanged.** | `a02e504ae56786d94f6a859e55b5c7e3229749ea3d43147424cfd987b2ec550e`, 4400931 bytes. |
| Zendesk Support | **Retained unchanged.** | `a487892c8e1f3feeba96c234148be69fddd50afce17bf30437bcb8de36d9a0c8`, 1757202 bytes. |

Independent local verification recomputed every retained artifact's byte count and SHA-256 and
matched it both to its new `*-retained-artifacts.json` record and to its parity-lock artifact
identity. The retained set is **10/20 connectors, 10 artifacts**: BambooHR, Fastly, HubSpot,
LaunchDarkly, Pipedrive, QuickBooks, ShipStation, Squarespace, WooCommerce, and Zendesk Support.
There are **two re-pins** (BambooHR, LaunchDarkly) and **eight preserved pins**. The unresolved
half is intentionally left without a raw artifact rather than retaining a changed document,
application shell, error page, or incomplete multi-artifact set.

## Source-cited deferred-operation audit — 2026-08-24

### Manual GSD / TDD plan

The canonical GSD paths for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` were re-resolved and their generated prompts reviewed for #4349.
This is the documented inline/manual fallback: the canonical connector delivery contract forbids
role spawning, and Firstmate expressly limited this slice to cited deferred-operation accounting
while transport admission remains external. Required skills loaded: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`. The CLI
help/manual/website parity guidance was read; no `pm` command, help, output, or generated manual
changes in this slice, so all runtime/manual/website parity edits are not applicable.

- **Scope:** source-citation accounting in the twenty existing declaration-disposition ledgers,
  their two stale re-pin bases, and binary credential-boundary evidence. No `cli_surface`,
  `operations.json`, `writes.json`, `sync_transport.json`, engine, or command-runner change.
- **Red:** the exact source-basis audit found BambooHR's disposition ledger still cited
  `1562871` / `ecfc6382…00ebf15a` and LaunchDarkly's still cited `2953979` /
  `41fc8c76…e5469419`, although their retained parity locks now establish
  `1627565` / `a04e6544…f1177890` and `2936808` / `5712a80c…aa8d9b3e`, respectively.
  `go run ./cmd/connectorgen source-import fastly --check` also fails before projection because
  this batch has no `fastly-operation-source-lock.json`; all twenty carry legacy parity locks,
  not v3 operation locks/descriptors. That blocks application of
  `sourceProjectionApplyNonExecutableMutationDispositions` without inventing a source contract.
- **Green:** correct only BambooHR and LaunchDarkly's ledger `source_basis` SHA-256 and bytes to
  their already retained, connector-owned parity-lock identities. Re-audit all ten retained
  connectors, validate the two edited bundles, and preserve the source projection refusal as an
  explicit foundation/source-lock-transition gap rather than a command downgrade or a fabricated
  deferred-operation descriptor.
- **Credential-boundary proof:** build `pm`; in isolated initialized projects with no credential
  configured, invoke every existing `availability: implemented` command. Success is only exit 1
  with exactly `error: missing --credential`; no provider call is possible at that boundary.

### Audited split before the narrow citation repair

The legacy declaration-disposition ledgers have **7,421 mapped source rows**. A row holds its
provider URL and location; its connector-level `source_basis` supplies the SHA-256 and byte count.
The table separates a currently usable source citation from a known source transition or a
no-public-source verdict. Thus a reported `unexplained` zero does not conceal a stale source URL:
those rows have their own named state and are not promoted as cited provider contracts.

| Connector | Mapped source rows | Credential-boundary commands | Cited-and-deferred | Source transition / unavailable | Unexplained |
| --- | ---: | ---: | ---: | ---: | ---: |
| Airtable | 105 | 0 | 105 | 0 | 0 |
| BambooHR | 345 | 0 | 345 | 0 | 0 |
| Buildkite | 132 | 0 | 132 | 0 | 0 |
| eBay Fulfillment | 11 | 0 | 0 | 11 unavailable | 0 |
| Fastly | 732 | 0 | 732 | 0 | 0 |
| Google Analytics Data API | 28 | 0 | 0 | 28 transition | 0 |
| HubSpot | 3,118 | 0 | 3,118 | 0 | 0 |
| LaunchDarkly | 399 | 0 | 399 | 0 | 0 |
| Linear | 543 | 0 | 0 | 543 transition | 0 |
| Mailchimp | 323 | 295 | 0 | 28 transition | 0 |
| Pinterest | 284 | 0 | 284 | 0 | 0 |
| Pipedrive | 218 | 0 | 218 | 0 | 0 |
| QuickBooks | 134 | 0 | 0 | 134 transition | 0 |
| Salesforce | 10 | 0 | 0 | 10 transition | 0 |
| ShipStation | 47 | 0 | 47 | 0 | 0 |
| SonarCloud | 157 | 0 | 157 | 0 | 0 |
| Squarespace | 53 | 0 | 53 | 0 | 0 |
| TikTok Marketing | 7 | 0 | 0 | 7 unavailable | 0 |
| WooCommerce | 140 | 0 | 140 | 0 | 0 |
| Zendesk Support | 635 | 95 | 540 | 0 | 0 |
| **Total** | **7,421** | **390** | **6,270** | **761** | **0** |

The five transition counts are Google Analytics Data API, Linear, Mailchimp, QuickBooks, and
Salesforce. Their actual replacement URLs and fetched byte counts are cited in
`data/cli-top100-source-research-r1/report.md` rows 31, 34, 36, 39, and 43, but that report does
not provide a checked-in replacement SHA-256/location projection. They remain source-transition
debt until a reviewed source-lock migration can supply all four citation values. eBay Fulfillment
and TikTok Marketing are the research report's explicit no-public-source verdicts (rows 49 and
42); they remain declared debt, not shell/error-page derived pseudo-citations.

All 390 existing implemented commands have now been proved against the built binary: Mailchimp
**295/295** and Zendesk Support **95/95** each stopped at exactly `error: missing --credential` in
isolated initialized projects. No credentials or provider requests were used. The remaining
7,031 rows are declared `partial`/`declaration-pending`: 6,270 are currently cited deferred debt,
743 are source-transition debt, and 18 have an explicit no-public-source verdict. None claims
that the runtime can execute it.

The new v3 source-projection disposition mechanism remains unavailable for this legacy batch until
each affected connector receives a verified `*-operation-source-lock.json` and generated
descriptor. It cannot be forced from a parity lock: `source-import` is intentionally hermetic and
refuses the absent v3 lock before projection. Seven connectors also lack a usable current source
artifact or provider source (Airtable, Buildkite, Google Analytics Data API, Linear, Mailchimp,
Pinterest, SonarCloud drift/volatile; eBay Fulfillment and TikTok Marketing no public source;
Salesforce tenant-dynamic), so a batch-wide v3 migration would require separate source-transition
work, not the still-pending transport-admission foundation. No such migration is attempted here.

### Green verification

- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceRetain|TestSourceProjectionSourceCitedNonExecutableMutationDispositionsCoverAbsentAndIncompleteActions'`
  passed.
- `go test -timeout 20m ./internal/connectors/defs` passed.
- `go build ./cmd/pm` passed before the credential-boundary sweep; each of the 390 invocations was
  run against that built binary, partitioned into isolated initialized temporary projects with no
  credential configured.
- `go run ./cmd/connectorgen validate internal/connectors/defs/bamboo-hr` and
  `go run ./cmd/connectorgen validate internal/connectors/defs/launchdarkly` both passed.
- An exact byte-count/SHA-256 audit passed for the ten retained connector artifacts and their
  corresponding disposition-ledger `source_basis` records; the two corrected bases now equal the
  retained lock identities.
- `go run ./cmd/agentcontractgen check`, `make docs-check`, and `git diff --check` passed.

The repository's aggregate `go test -timeout 20m ./...` and `make verify` are deliberately left
to CI for this planning/provenance-only checkpoint: repository guidance warns that the full
550-plus-connector suite exceeds the per-command memory/time window. The focused affected-package
checks above use the required explicit timeout.

## Captain correction: full operation map and declaration basis — 2026-08-24

### Re-measurement and binary proof

After merging current `origin/main` commit `060bb7864e` (the binary-upload surface merge), the
independent whole-tree command inventory is: **2,191 reverse-ETL commands across 27 connectors**,
**2,002 direct reads across 21**, **645 ETL commands across 34**, **283 direct writes on GitHub
only**, **27 binary downloads across six**, and **one binary upload**. The changing reverse-ETL
total is branch-local conversion progress; it is not a missing engine capability. No connector
specific engine or shared-code change is authorized for this batch.

The 390 Batch 4/5 runnable operations were remeasured rather than inferred from their manifest:

- built the binary with `go build -o pm ./cmd/pm`;
- enumerated every `availability: implemented` command from the two actual surfaces (Mailchimp
  295, Zendesk Support 95);
- split the list into 5-command shards, each using a fresh `mktemp -d` project initialized with
  `pm init --root <project> --json` and no configured credential;
- invoked `pm <connector> <command path>` for every command and required exit status 1 and the
  complete combined output `error: missing --credential` exactly.

The final corrected sweep passed **295/295 Mailchimp** and **95/95 Zendesk Support** on the binary
built after the `060bb7864e` merge. Two early scratch harness attempts are explicitly discarded:
one used an invalid jq range selector and processed no command; one shifted away the first argument
in each shard. Neither contributes to the count. The corrected `pipefail`-protected sweep emitted
only 5-command pass shards, totalling 390, with no provider request or credential. Thus the first
category below is an observed binary result, not a declaration count.

### Three-way source operation split — report before conversion

| Category | Source operations | Meaning |
| --- | ---: | --- |
| **Already runnable** | **390** | The built-binary credential-boundary result above (Mailchimp 295; Zendesk Support 95). |
| **Declarable now** | **5,543** | The operation has a retained, connector-owned provider artifact and no known explicitly open request map. It can be expressed entirely in that connector's existing definition files as a typed action/stream/command; no shared runtime change is permitted or needed. |
| **Genuinely blocked** | **1,488** | 1,439 lack a current retained provider contract for field derivation; 49 have a provider-declared open request map and cannot honestly be represented by bounded flags. |
| **Total** | **7,421** | Every legacy mapped source operation is in exactly one category. |

The conversion-ready count is per connector: BambooHR 345, Fastly 732, HubSpot 3,082, LaunchDarkly
399, Pipedrive 218, ShipStation 47, Squarespace 53, WooCommerce 140, and Zendesk Support 527.
Those totals are the retained source inventories, less the 95 already runnable Zendesk operations
and the known open request maps described below. This is a large declaration wave, not an engine
foundation task.

Blocked evidence is explicit rather than a catch-all:

| Block reason | Operations | Connector evidence |
| --- | ---: | --- |
| Source artifact has no retained current bytes | 678 | Airtable 105, Buildkite 132, Pinterest 284, SonarCloud 157. Each operation retains its historical provider URL/location in its declaration-disposition ledger, but the current dynamic rendered reference was not retained as an immutable source contract. |
| Wrong/transition source identity | 743 | Google Analytics Data API 28, Linear 543, Mailchimp 28, QuickBooks 134, Salesforce 10. Their provider replacement URLs and fetched byte counts are evidenced in the source-research report rows 31, 34, 36, 39, and 43; a replacement connector-owned SHA-256 and operation-location projection is not yet present, so they are not promoted as cited executable contracts. |
| Provider source retrieval failed | 18 | eBay Fulfillment 11 (provider OpenAPI URL returned HTTP 403, 1,831-byte non-source) and TikTok Marketing 7 (provider documentation TLS failure, no body), as recorded by source-research rows 49 and 42. A failed fetch is evidence of the retrieval block, not a claim the provider has no documentation. |
| Provider explicitly permits open request fields | 49 | HubSpot 36 and Zendesk Support 13 existing `rest_write` body schemas contain an open object (`additionalProperties` not `false`). No bounded flags or invented closed schema may be authored for these operations. |

For every source row, the authoritative citation remains its connector's
`sources/<connector>-declaration-disposition.json`: `ledger_dispositions[*].source` supplies the
provider URL and operation location, while `source_basis` supplies the lock URL, SHA-256, and byte
count. The nine conversion-ready connectors additionally retain the exact artifact and provenance
manifest under `sources/artifacts/` and `sources/<connector>-retained-artifacts.json`. The
transition and retrieval-block rows above are deliberately **not** called source-cited executable
contracts; their cited research evidence explains precisely why they cannot be declared yet.

The 5,543 declarations require connector-local `writes.json`/`streams.json`/`cli_surface.json`
and existing companion declaration updates only. Where a source request shape resolves to another
open map during conversion, it moves from the second category to the fourth-row open-map reason
with its per-operation source citation; it is never forced through a JSON catch-all or a shared
connector special case.

Post-merge report checks passed: `go build -o pm ./cmd/pm`; the 390-command current-binary sweep
above; `go test -timeout 20m ./internal/connectors/defs`; `make docs-check`; and
`go run ./cmd/agentcontractgen check`.

## Conversion red gate — operation identities are not typed executable contracts — 2026-08-24

### Correction to the reported conversion count

The reported **5,543** was not an operation count from the providers' retained artifacts. It is a
legacy declaration-mapping count which includes **39 `local_execution_bindings`**: BambooHR 26,
LaunchDarkly 2, Pipedrive 5, and Zendesk Support 6. Those bindings are local implementation
references, not provider operations, and cannot become provider-cited commands. The corrected
number of retained, provider REST operation identities remaining after the already-runnable
Zendesk rows and the 49 known explicit open request maps is **5,504**:

| Connector | Retained provider REST operations | Non-provider local bindings previously included | Deduction | Candidate provider operations |
| --- | ---: | ---: | ---: | ---: |
| BambooHR | 319 | 26 | 0 | 319 |
| Fastly | 732 | 0 | 0 | 732 |
| HubSpot | 3,118 | 0 | 36 known open request maps | 3,082 |
| LaunchDarkly | 397 | 2 | 0 | 397 |
| Pipedrive | 213 | 5 | 0 | 213 |
| ShipStation | 47 | 0 | 0 | 47 |
| Squarespace | 53 | 0 | 0 | 53 |
| WooCommerce | 140 | 0 | 0 | 140 |
| Zendesk Support | 629 | 6 | 95 already runnable; 13 known open request maps | 521 |
| **Total** | **5,648** | **39** | **144** | **5,504** |

This fixes the number without downgrading a command: all 390 existing credential-boundary
observations remain intact, and every remaining row stays partial until a source-derived executor
contract is present.

### Red evidence: no safe connector-local promotion exists from this inventory alone

`source-import` is intentionally operation-lock-only. The batch has only legacy
`*-parity-source-lock.json` files, so it refuses before source projection when the required
`*-operation-source-lock.json` is absent (the prior exact `fastly --check` failure is retained
above). This is correct: the parity lock records operation identity and artifact provenance, not
the typed request parameters, body media/schema, output envelope, or executable bounds used by
the runtime.

Even after an operation descriptor exists, the current projector deliberately does **not** create
a write action, direct-read operation, stream schema, fixture, or command route from an endpoint.
`cmd/connectorgen/sourceprojection.go` matches a source operation only to an already-authored
connector action and returns `source operation(s) have no complete executable action` for the
rest. That guard is the reason an operation count cannot be promoted by changing
`availability`. It preserves the no-generic-body and no-invented-`additionalProperties:false`
requirements.

The retained artifacts therefore divide by contract form, not merely by connector count:

| Connector | Retained source form | Provider-operation identities | Existing definition gap that prevents a binary-reachable command |
| --- | --- | ---: | --- |
| BambooHR | Rendered HTML reference | 319 | No operation lock, typed operation/action contract, or command bindings; 26 local bindings are not provider operations. |
| Fastly | Postman collection JSON | 732 | The collection supplies request examples, not a parseable OpenAPI `paths` contract for the existing importer; no typed actions or operation routes exist. |
| HubSpot | Gzip archive containing 524 JSON documents | 3,118 | The importer accepts a retained source document, not an archive corpus; the current metadata-only `operations.json` does not supply executable action/flag contracts. |
| LaunchDarkly | OpenAPI 3.0.3 | 397 | No operation lock, actions, streams, or operation routes. |
| Pipedrive | OpenAPI 3.0.1 | 213 | Seventeen legacy actions exist but none has a CLI binding; the other provider operations have no action or direct-read route. |
| ShipStation | OpenAPI 3.1.0 | 47 | No actions or operation routes. |
| Squarespace | OpenAPI 3.1.1 | 53 | Two legacy actions exist but no CLI bindings; the other provider operations have no action or direct-read route. |
| WooCommerce | Rendered HTML reference | 140 | No operation lock, typed operation/action contract, or command bindings. |
| Zendesk Support | OpenAPI 3.0.3 | 629 | Ninety legacy actions exist; 62 are already runnable, while 28 remain partial and the remaining source operations lack action/direct-read contracts. |

Existing legacy actions are not a safe shortcut. For example, the partial Zendesk create/update
actions contain provider-declared nested open objects such as `automation`, `group`, `ticket`,
and `user`; changing only their command availability would make the binary stop at credentials
while leaving the action body less bounded than this task permits. Conversely, a `direct_read`
row needs either a declared stream or an exact `operations.json` operation and its input/output
bindings; a source URL and `api_surface` endpoint are insufficient under the command-runner
preflight.

### Manual GSD/TDD decision point

- **Red:** the corrected conversion map has 5,504 candidate provider operation identities but
  zero new typed executable contracts. Attempting to promote any of them from only the parity
  lock would violate both the runtime's source-projection guard and the task's provider-fidelity
  rule.
- **Green condition:** for a selected source form, derive and retain an operation lock/descriptor,
  then author connector-owned typed action, direct-read operation, or stream contracts only where
  the source provides every required field, body, response, pagination, and bound. Run
  `surface-sync`, bundle validation, and a fresh built-binary credential-boundary probe for every
  promoted command.
- **Refactor/guard:** do not add a generic JSON body, close an open provider object, or reclassify
  a local execution binding as a provider operation. A source row whose body or response remains
  open stays partial with its existing citation.

This is a delivery-scope decision, not a connector-local implementation detail. Reaching the
green condition for the OAS connectors needs a reusable source-lock-to-action/operation authoring
path; the HTML, Postman, and HubSpot archive forms additionally need source-form-specific
extraction before any fields can be declared. The task prohibits adding that shared foundation in
this connector-only wave. The safe next authority is either (1) a separately scoped foundation
issue for artifact-corpus/source-contract extraction followed by connector conversion waves, or
(2) an explicitly narrowed connector-local wave with a named set of provider operations and
individually retained typed contracts. No availability value has been changed pending that choice.

## Captain-directed individually-sourced split — pre-conversion report — 2026-08-24

The delivery owner selected the individually-sourced route: do **not** add a shared
source-contract extraction foundation and do **not** rewrite a source lock. This replaces the
foundation decision point above. The figures below classify the corrected **5,504 provider REST
operation identities**, not the 39 local execution bindings and not the 390 commands already
proven at the credential boundary.

| Connector | A — retained source contains field-level machine contract | B — retained source is identity/examples; provider docs must be captured per operation | C — neither source form is available | Total |
| --- | ---: | ---: | ---: | ---: |
| BambooHR | 319 | 0 | 0 | 319 |
| Fastly | 0 | 732 | 0 | 732 |
| HubSpot | 3,082 | 0 | 0 | 3,082 |
| LaunchDarkly | 397 | 0 | 0 | 397 |
| Pipedrive | 213 | 0 | 0 | 213 |
| ShipStation | 47 | 0 | 0 | 47 |
| Squarespace | 53 | 0 | 0 | 53 |
| WooCommerce | 0 | 140 | 0 | 140 |
| Zendesk Support | 521 | 0 | 0 | 521 |
| **Total** | **4,632** | **872** | **0** | **5,504** |

### Category A — individually declarable from the retained provider contract (4,632)

This category means that the **retained bytes themselves** contain provider request/response
field definitions. It is not a claim that toggling `availability` is valid: every operation still
needs its own connector-local stream/action/operation and CLI declaration, source location,
bounded body/input mapping, and built-binary credential-boundary proof.

- BambooHR's retained rendered reference embeds an OpenAPI 3.1 document (`"openapi":"3.1.0"`)
  and records exact `paths.<path>.<method>` locations for all 319 provider operations in
  `sources/bamboo-hr-declaration-disposition.json`.
- HubSpot's retained gzip corpus, LaunchDarkly's OpenAPI 3.0.3, Pipedrive's OpenAPI 3.0.1,
  ShipStation's OpenAPI 3.1.0, Squarespace's OpenAPI 3.1.1, and Zendesk Support's OpenAPI 3.0.3
  are provider-published machine contracts. Their immutable byte identities and public source URLs
  are in each connector's `sources/<connector>-retained-artifacts.json`; each operation's
  provider location is in its `sources/<connector>-declaration-disposition.json` ledger.
- The count excludes the 36 already-known HubSpot and 13 already-known Zendesk operations whose
  provider request contract is an open object, as well as the 95 already-runnable Zendesk
  operations. Those 144 deductions are retained exactly as in the corrected count table above;
  no operation gains a fabricated `additionalProperties: false` declaration.

An exploratory read of the existing generic importer against LaunchDarkly stops at an OpenAPI
3.0 reference with an `example` sibling. That is a current importer limitation, not evidence that
the provider omitted fields. It confirms the captain's direction: this wave must author an
individual static declaration from the cited provider contract or leave that operation blocked;
it must not change the importer or silently discard the sibling.

### Category B — provider documentation required before each declaration (872)

Fastly's retained Postman collection has request/response **examples**, not an authoritative
complete field contract for every operation; WooCommerce's retained HTML indexes operations but
does not provide an immutable per-operation typed capture in this bundle. Both providers publish
their own API references:

- Fastly: <https://www.fastly.com/documentation/reference/api/> (732 retained operation
  identities; current retained collection digest
  `c6ae5b0fd118fe2d87e7d0ef431f67cda703d1487d27b5f02725d82219386677`, 2,341,028 bytes).
- WooCommerce: <https://developer.woocommerce.com/docs/apis/rest-api/v3/> (140 retained
  operation identities; current retained rendered reference digest
  `a02e504ae56786d94f6a859e55b5c7e3229749ea3d43147424cfd987b2ec550e`, 4,400,931 bytes).

Before a Category-B operation is converted, the specific official reference page that defines its
request and response fields must be captured as connector-owned evidence with its public URL,
SHA-256, byte count, and page/section location. That evidence is additive; it does **not** alter
either existing lock. Until then, all 872 remain partial rather than being promoted from a request
example or an operation name.

### Category C — no source contract (0 of the 5,504 candidates)

There are no Category-C rows among the corrected candidate set: each candidate has either a
retained field-level provider contract (A) or a provider documentation route identified above
(B). This does not erase the earlier 1,439 source-unavailable/wrong-identity/retrieval-blocked
operations: they are outside the corrected 5,504 candidate set and remain blocked with the
citations already recorded in this trace.

### Manual GSD/TDD checkpoint

- **Red:** the legacy partial commands are endpoint declarations, not executable contracts; the
  5,504 total cannot be promoted by metadata alone.
- **Green condition for A:** a connector-local source-derived contract is authored for a named
  operation and the built `pm` binary exits at exactly `error: missing --credential` in its own
  no-credential project.
- **Green condition for B:** first add the additive official-document evidence (URL, SHA-256,
  bytes, and source location), then meet the same contract and binary proof.
- **Guard:** no source-lock rewrite, no shared engine/importer change, no generic JSON body, no
  invented closed object, and no availability-only promotion.

No conversion has begun. This is the required count report before any bulk declaration change.

## Wave 1 — bounded, high-variance credential-boundary proof plan — 2026-08-24

The delivery owner approved a **100-operation Category-A Wave 1**, rather than promotion of all
4,632 source-contract candidates. This is an intentionally failure-seeking sample: a repeated
contract/dispatch error is useful at 100 and must stop the wave before any larger conversion.

| Intent | Count | Connector spread | Contract variety exercised |
| --- | ---: | --- | --- |
| `direct_read` | 0 | Set aside | The validator rejects every Category-A direct-read endpoint because no endpoint is covered by an executable stream or operation contract. Promoting one would be an availability-only claim, so this lane remains partial. |
| `etl` | 100 | BambooHR 62; LaunchDarkly 5; Pipedrive 21; ShipStation 4; Squarespace 8 | Each command is joined exactly to a pinned GET endpoint already covered by its named connector-owned stream, including nested/array-bearing schemas and both collection and configured-resource routes. |
| `direct_write` | 0 | Set aside | The three earlier HubSpot candidates remain typed metadata, not executable direct-write contracts; the destination-transport gap recorded below also applies to the write lane. |
| **Total** | **100** | **7 connectors** | **No source lock or engine change.** |

### Reverse-ETL exception recorded before conversion

The requested intent spread cannot safely include a newly promoted `reverse_etl` command. None of
the seven Wave-1 connectors has `sync_transport.json`; the existing partial Zendesk Support write
rows also include the thirteen source-declared open request objects already excluded from Category
A. Adding a connector-local transport solely to make one test command pass would invent a
destination contract, and changing the shared transport model is out of scope. Wave 1 therefore
records this as an intentional, source-cited exception rather than mislabelling a direct write as
reverse ETL. It is not a downgrade of any of the 390 commands already proven on `main`.

### Red evidence

From a newly initialized, credential-free project, the freshly built current binary rejects the
first selected BambooHR ETL command before the credential boundary:

```text
pm bamboo-hr operations get-api-v1-employees-directory --root <fresh-root> --json
exit 1
error: missing --credential
```

The source endpoint and stream binding are already provider-derived in each connector's
`api_surface.json`, `streams.json`, source lock, and declaration-disposition ledger. Green work
adds that exact existing stream name to the corresponding partial ETL command, then proves the
built binary's exact terminal result:

```text
error: missing --credential
```

### Green/stop rules

- Build `pm` after the declaration edits and run every one of the 100 commands from a fresh,
  credential-free initialized project. A command counts only if its combined output is exactly the
  `missing --credential` terminal error; no provider request is permitted.
- The first preflight, field-binding, output-policy, or dispatch failure is classified by source
  operation and stops the wave. Do not compensate by loosening a schema, suppressing a flag, or
  changing shared code.
- Category B (Fastly and WooCommerce) remains untouched until this entire Wave 1 report is green.

### Wave-1 write-lane exception — 2026-08-24

The planned three HubSpot `direct_write` probes cannot be promoted honestly. Each selected
operation is retained as typed metadata, but its connector-owned `operations.json` explicitly
says it “is not executable until a future connector-local implementation covers it.” The runtime
also requires a connector to implement `OperationDirectWriter`, direct-write metadata,
declaration binding, and body-materialization interfaces before an `implemented` direct-write
command may enter the plan lifecycle (`commandrunner.validateOperationDirectWriteCommand`).
Neither condition is met by changing `cli_surface.json`, and constructing either missing contract
would violate the Wave-1 no-foundation/no-invented-body constraint.

The reverse-ETL exception remains unchanged: none of the 20 target connectors owns a
`sync_transport.json`. Therefore the initial four-intent sample cannot proceed as specified. This
is a pre-conversion evidence finding, not a failed conversion. The delivery owner has expressly
set the 390 write actions aside and authorized the independent 100-command read/ETL Wave 1 above;
all write declarations remain unchanged.

### Transport registration check — 2026-08-24

The existing `declarative_api` **source** executor is generic: production composition scans every
definition, registers the exact `declarative_stream_source` reference without selecting a connector
by name, and admits an engine-backed connector whose `eligible_streams` exactly covers its declared
streams. A connector-local source role can therefore use that executor without an engine change.

The existing `declarative_api` **destination** executor is also generic in implementation, but it
is not usable from `sync_transport.json` alone. The closed
`declarative_typed_destination` adapter refuses admission unless the connector has, for every
selected action, a provider-owned idempotency header, an action-owned bounded source binding, and
a provider read-back contract coupled to its conformance evidence. Those are required to prevent a
replayed mutation or a local receipt from posing as provider acknowledgement.

The eight target bundles that contain `writes.json` have 390 actions in aggregate (Airtable 12,
BambooHR 101, Buildkite 17, Mailchimp 148, Pipedrive 17, SonarCloud 8, Squarespace 2, Zendesk
Support 90); **zero** declares `idempotency_key_header` and **zero** declares the required
read-back contract. The other twelve target bundles have no write actions. This is a real missing
provider-contract boundary, not a missing generic executor: writing only `sync_transport.json`
would fail destination registration before credential resolution. No transport file or command
surface was changed.

### Wave-1 green report — 2026-08-25

**Correctly mapped and proven: 100 ETL commands.** The mapping criterion was exact and
connector-local: a partial ETL command's `summary` had to equal `Declared etl: GET <path>.` for a
pinned `api_surface.json` `GET <path>` endpoint whose `covered_by.stream` named the exact existing
stream. The command then gained only that stream name, `availability: implemented`, and a binding
note. No source lock, stream schema, endpoint, shared engine, or transport declaration changed.

| Connector | Exact endpoint-to-stream bindings | Built-binary probes | Result |
| --- | ---: | ---: | --- |
| BambooHR | 62 | 62 | 62 exact `error: missing --credential` |
| LaunchDarkly | 5 | 5 | 5 exact `error: missing --credential` |
| Pipedrive | 21 | 21 | 21 exact `error: missing --credential` |
| ShipStation | 4 | 4 | 4 exact `error: missing --credential` |
| Squarespace | 8 | 8 | 8 exact `error: missing --credential` |
| **Total** | **100** | **100** | **100/100 proven** |

The binary was rebuilt after the declaration changes. Each probe ran sequentially from its own
newly initialized project with no configured credential; each exited `1`, printed the terminal
line `error: missing --credential`, and did not print `unknown command` or make a provider call.
The 100 checks were bounded into 10-or-fewer-command batches only to stay below the execution
window; this did not share credentials or execution state.

**Partial/deferred: direct reads.** The attempted high-variance direct-read sample was not
promoted. `connectorgen validate` rejects an `implemented` direct read unless its endpoint is
covered by an executable stream/operation contract. The source-defined partial command for
`GET /cms/site-search/2025-09/search`, for example, fails as “not covered by an executable
surface.” The same absence applies to the Wave-1 direct-read candidates across the seven planned
connectors, so they remain partial rather than becoming an availability-only claim.

**Unsupported/set aside: writes.** The 390 write actions remain set aside under the previously
recorded destination-transport foundation gap: zero source-declared idempotency headers and zero
action-owned provider read-back contracts. No write or reverse-ETL command was promoted or
retested in this wave.

### Next bounded increment — remaining exact BambooHR ETL bindings

- **Red:** 22 source-cited BambooHR `GET` operation rows still have a partial ETL command despite
  an exact `api_surface.json` `covered_by.stream` match; they are the remaining rows after Wave 1
  selected the first 62 of 84 exact BambooHR endpoint-to-stream bindings. The earlier 83 count was
  a selection-query undercount; the pre-wave bundle confirms 84 exact partial candidates and zero
  already implemented candidates.
- **Green condition:** bind only the exact existing stream, validate the bundle and generated
  surface, then run all 22 built-binary credential-free probes sequentially. Any result other than
  terminal `error: missing --credential` stops this increment; direct reads and writes remain out
  of scope.

### Next bounded increment — green report

**Correctly mapped and proven: 22 BambooHR ETL commands.** Each command already cited its pinned
provider operation. The same exact connector-local mapping criterion used in Wave 1 applied: its
`Declared etl: GET <path>.` summary matched a `GET <path>` endpoint in `api_surface.json`, and that
endpoint's `covered_by.stream` named the existing stream. Only the named stream, implemented
availability, and a binding note changed.

The rebuilt binary ran the 22 commands from fresh initialized, credential-free projects in three
bounded sequential batches (8, 8, and 6). All 22 exited `1` with exactly the terminal line
`error: missing --credential`; none printed `unknown command` or contacted a provider. BambooHR
now has 84 source-cited, endpoint-to-existing-stream ETL bindings. Together with Wave 1, the
honest exact-binding total is 122 commands: BambooHR 84, LaunchDarkly 5, Pipedrive 21, ShipStation
4, and Squarespace 8. Direct reads and write lanes remain unchanged.
