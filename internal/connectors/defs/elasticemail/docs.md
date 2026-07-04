# Overview

Elastic Email is a wave2 fan-out declarative-HTTP migration, expanded to the full documented API
surface in Pass B (2026-07-04). It reads contacts, campaigns, lists, segments, templates, domains,
suppressions (all/bounces/complaints/unsubscribes), webhooks, files, inbound routes, sub-accounts,
and campaign/channel statistics through the Elastic Email v4 REST API (`{{ config.base_url }}/...`),
and writes contact/list/segment/template/campaign/webhook/domain/inbound-route lifecycle mutations.
This bundle originated as a migration from `internal/connectors/elasticemail` (the hand-written
connector it replaces, which stays registered and unchanged until wave6's registry flip);
`capabilities.write` is now `true` following the Pass B write-surface expansion — this is a
capability WIDENING beyond legacy's read-only `Write` stub, not a parity port, since Elastic Email's
real API genuinely exposes these mutation endpoints and legacy never called them.

## Auth setup

Provide an Elastic Email API key via the `api_key` secret; it is sent as the
`X-ElasticEmail-ApiKey` header with no prefix (`mode: api_key_header`, empty `prefix`), matching
legacy's `connsdk.APIKeyHeader("X-ElasticEmail-ApiKey", secret, "")`. `base_url` defaults to
`https://api.elasticemail.com/v4` and may be overridden for test proxies.

## Streams notes

All 5 streams share the identical shape: `GET`, records at the response root (`records.path: ""`
— every Elastic Email v4 list endpoint returns a top-level JSON array, matching
`connsdk.RecordsAt(resp.Body, "")`'s root-array selection and legacy's own `RecordsAt(resp.Body,
"")` call), and `offset_limit` pagination (`limit`/`offset` query params). Primary keys match each
stream's natural identifier rather than a synthetic id, exactly as legacy's own catalog declares:
`Email` for `contacts`, `ListName` for `lists`, `Name` for `campaigns`/`segments`/`templates`.
`contacts` declares a bare `incremental.cursor_field: "DateUpdated"` (no `request_param`) and its
schema's `x-cursor-field` names the same field, matching legacy's `Catalog()` `CursorFields:
[]string{"DateUpdated"}` hint for the `contacts` stream — legacy's own `Read` never filters on it
(Elastic Email v4 list endpoints have no request-side time-range filter parameter its own
connector code ever sends), so per `docs/migration/conventions.md`'s incremental-declaration truth
table (bare `cursor_field` iff legacy publishes `CursorFields`; `request_param` iff legacy sends a
server-side filter) this is a bare declaration only, not a `request_param`. `campaigns`/`lists`/
`segments`/`templates` declare no `CursorFields` in legacy and correspondingly have no
`incremental` block or `x-cursor-field`.

`contacts`' schema includes `Activity`/`Consent`/`CustomFields` (nested objects legacy's
`contactRecord` mapper emits) even though legacy's separate `contactFields()` catalog-description
function omits them — the schema is a projection of what `mapRecord` actually emits (the
authoritative behavior per `docs/migration/conventions.md`'s schema-as-projection rule), not of
the narrower `Fields` catalog list, which is descriptive metadata only and does not gate what
`Read` returns.

**Pass B new streams** (all `GET`, root-array records, `offset_limit` pagination unless noted):
`domains` (`GET /domains`, `pagination: none` — the real endpoint accepts no `limit`/`offset` query
params at all, matching `DomainsGet`'s documented zero-parameter shape), `suppressions` /
`suppressions_bounces` / `suppressions_complaints` / `suppressions_unsubscribes` (4 distinct
endpoints returning the same `Suppression` record shape for different suppression categories),
`webhooks` (`GET /webhook`, singular path segment per the real API), `files`, `inbound_routes`
(`GET /inboundroute`, `pagination: none` — no `limit`/`offset` params documented), `sub_accounts`,
`statistics_campaigns` (`GET /statistics/campaigns`) and `statistics_channels` (`GET
/statistics/channels`, same `ChannelLogStatusSummary` record shape as `statistics_campaigns`). None
of these declare an `incremental` block — the real API documents no server-side date-range filter
parameter for any of them (mirroring the existing `campaigns`/`lists`/`segments`/`templates`
streams' own full-refresh-only shape).

## Write actions & risks

Pass B added `capabilities.write: true` and 22 actions across 8 resources — see `writes.json` for
the full `record_schema`/`risk` text per action. Summary by resource:

- **contacts**: `create_contact` (`POST /contacts`), `update_contact` (`PUT /contacts/{Email}`),
  `delete_contact` (`DELETE /contacts/{Email}`, idempotent on 404).
- **lists**: `create_list`, `update_list` (`PUT /lists/{ListName}`, body restricted to
  `NewListName`/`AllowUnsubscribe` — the real `ListUpdatePayload` shape, distinct from the create
  payload), `delete_list`, `add_list_contacts` (`POST /lists/{ListName}/contacts`).
- **segments**: `create_segment`, `update_segment`, `delete_segment`.
- **templates**: `create_template`, `update_template`, `delete_template`.
- **campaigns**: `create_campaign` (requires `Name` + `Recipients`; flagged **higher-risk** — a
  campaign create/update can schedule a live send to real recipients depending on `Options`, this
  is not a preview-only action), `update_campaign`, `pause_campaign` (`PUT
  /campaigns/{Name}/pause`), `delete_campaign`.
- **webhooks**: `create_webhook`, `update_webhook` (path-keyed on `WebhookID`, the real API's
  identifier — NOT `Name`), `delete_webhook`. Flagged **higher-risk**: registers/repoints a
  caller-controlled external URL that receives live send-event data.
- **domains**: `create_domain`, `delete_domain`. Update/verification-trigger endpoints are excluded
  (`requires_elevated_scope` — see `api_surface.json`) since they depend on externally-configured
  DNS state this connector cannot validate.
- **inbound_routes**: `create_inbound_route`, `update_inbound_route`, `delete_inbound_route`
  (path-keyed on `PublicId`). Flagged **higher-risk**: registers/repoints a caller-controlled
  external address/URL that receives forwarded inbound mail.

Every write action's `path_fields` names the real API's own path-parameter identifier (`Email` for
contacts, `ListName`/`Name` for lists/segments/templates/campaigns, `WebhookID` for webhooks,
`PublicId` for inbound routes, `Domain` for domains) — these are natural business keys, not
synthetic ids, matching how each resource's own list-stream schema names its `x-primary-key`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional
  `page_size` (1-1000, default 100) and `max_pages` (default unlimited, `all`/`unlimited`/`0`
  synonyms) config keys read at request time (`elasticEmailPageSize`/`elasticEmailMaxPages`). The
  engine's `PaginationSpec.PageSize`/`MaxPages` fields are plain fixed JSON integers baked into
  `streams.json` — there is no templating/config-driven override mechanism for them.
  `base.pagination.page_size` is set to legacy's real production default, `100`
  (`elasticEmailDefaultPageSize`) — this is the actual value every paginated live stream sends;
  it is not a fixture convenience. The required two-page `contacts` conformance fixture therefore
  records a full 100-row first page and a short second page, matching legacy's request cadence.
  No `max_pages` cap is declared (unbounded, matching legacy's own default). Neither key is
  declared in `spec.json` (F6, `docs/migration/conventions.md`: dead, unwireable config is worse
  than absent config).
- **`base_url` scheme/host validation is enforced by legacy in Go** with dedicated error messages
  (`elasticEmailBaseURL`); the engine has no equivalent declarative URL-shape validator, so a
  malformed `base_url` here surfaces as a generic request-construction/connection error rather
  than legacy's specific `"config base_url must use http or https"`/`"must include a host"`
  messages. This never changes behavior for any valid `base_url`.
- **Pass B full-surface review (2026-07-04).** `api_surface.json` now enumerates the entire
  documented Elastic Email v4 REST surface (~85 endpoints across 16 resource groups). Deliberately
  excluded, never modeled: `security/apikeys` and `security/smtp` credential management (the create
  response returns the raw secret key/password material itself — routing that through this
  connector's record/fixture paths is a credential-exfiltration risk, not merely out-of-scope
  breadth); every asynchronous export/import job endpoint (contacts/events/suppressions bulk
  import/export — kickoff-and-poll shapes, not synchronous record reads/writes); the raw/
  transactional email-send endpoints and the email-verification subsystem (both independent
  features outside contact/campaign/list data management); and any endpoint whose request or
  response body is binary/multipart (`binary_payload` — file upload/download, verification file
  upload). See `api_surface.json`'s per-endpoint `excluded.reason` for the specific justification
  behind every other omitted endpoint (mostly `duplicate_of` single-object detail lookups already
  covered by their list stream, and a handful of `requires_elevated_scope` account-administration
  writes — sub-account provisioning/credit adjustment, domain DNS re-verification).
- **`update_list`'s real request body is asymmetric with its create payload**: the API's
  `ListUpdatePayload` accepts `NewListName`/`AllowUnsubscribe` only (no `Emails` field, unlike
  `ListPayload`'s create-time seed list) — `update_list`'s `body_fields` is restricted accordingly
  so a caller-supplied `Emails` value on an update record is silently dropped from the request
  body rather than sent to a field the real endpoint does not accept.
