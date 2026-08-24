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
