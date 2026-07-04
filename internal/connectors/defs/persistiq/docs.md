# Overview

PersistIQ is a wave2 fan-out declarative-HTTP migration, expanded in Pass B beyond legacy parity.
It reads PersistIQ leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events,
lead fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies through v1 REST
endpoints (`GET https://api.persistiq.com/v1/...`), and now also creates/updates leads and
campaigns, adds/removes campaign leads, replies to campaign messages, and adds DNC domains. The 6
original streams (`leads`, `users`, `campaigns`, `mailboxes`, `activities`, `accounts`) are
engine-vs-legacy parity-tested against `internal/connectors/persistiq` (the hand-written connector
this bundle migrates); the legacy package stays registered and unchanged until wave6's registry
flip. Every Pass B addition (8 new streams, 7 write actions) has no legacy counterpart and carries
no parity constraint.

This bundle's Pass B research is grounded directly in PersistIQ's own live, machine-readable
OpenAPI 3.0.1 spec: `https://persistiq.com/api-docs/index.html` loads a Swagger UI that fetches
`https://persistiq.com/api-docs/v1/swagger.json` (14 documented paths), both confirmed reachable
by direct fetch during this review. **This corrects a prior review's assessment that this docs host
was dead/parked** — it is not; the earlier note was mistaken (see Known limits).

## Auth setup

Provide a PersistIQ API key via the `api_key` secret; it is sent as the raw `X-API-KEY` header
(`auth.mode: api_key_header`, `header: X-API-KEY`, no prefix) and is never logged, matching
legacy's `connsdk.APIKeyHeader("X-API-KEY", key, "")` (`persistiq.go:127`) — confirmed against the
live spec's own `securitySchemes.api_key` (`type: apiKey`, `name: x-api-key`, `in: header`; HTTP
header names are case-insensitive, so `X-API-KEY`/`x-api-key` are the same header). `base_url`
defaults to `https://api.persistiq.com` (the live spec's own `servers[0].url`) and may be overridden
for tests/proxies.

## Streams notes

All 6 legacy-parity streams (`leads`, `users`, `campaigns`, `mailboxes`, `activities`, `accounts`)
share the same base shape: `GET` against the PersistIQ v1 list endpoint, records at a top-level key
matching the stream name, primary key `["id"]`. Pagination is page-number based
(`pagination.type: page_number`, `page_param: page`, `size_param: per_page`, `start_page: 1`,
`page_size: 100`), stopping on a short page — matching legacy's default request shape
(`page=1`, `per_page=100`) and short-page stop. Legacy can override page size at runtime, but the
declarative pagination block is fixed at the legacy default; see Known limits.
This bundle keeps legacy's page-number shape unchanged even though the live spec's own list-response
envelope also carries `has_more`/`next_page` fields (see Known limits — a page-number-vs-cursor
pagination-shape reconciliation was judged out of scope for a parity stream without live-credential
verification).

No legacy stream declares an `incremental` block. Legacy always performs a full refresh for every
stream and does not publish cursor fields, so this bundle does the same even though `GET /v1/leads`
documents optional timestamp filters.

All 6 legacy streams declare `"projection": "passthrough"` (post-wave2 review §8 rule 1): legacy's
`Read` emits `emit(connectors.Record(rec))` — a verbatim type-cast of the raw harvested record,
with no `mapRecord`-style field-building — so schema-mode projection would silently drop any raw
field this bundle's schema omits.

**Pass B additions** (no legacy counterpart; authored directly from the live OpenAPI spec):

- `dnc_domains` (`GET /v1/dnc_domains`, records at `dnc_domains`): the account's Do-Not-Contact
  domain list.
- `events` (`GET /v1/events`, records at `events`): an audit-trail event stream (`event_type`,
  `data`, `created_at` per event).
- `lead_fields` (`GET /v1/lead_fields`, records at `lead_fields`): custom lead-field definitions.
- `lead_statuses` (`GET /v1/lead_statuses`, records at `lead_statuses`): the account's configured
  lead status values.
- `tags` (`GET /v1/tags`, records at `tags`): the account's lead tags.
- `webhook_plugin` (`GET /v1/webhook_plugin`, single-object response, `records: {"path": ".",
  "single_object": true}`, `pagination: {"type": "none"}` at the stream level since this endpoint
  is a settings singleton with no list/page semantics at all): the account's webhook-notification
  configuration (which events post to which URLs).
- `campaign_leads` (`GET /v1/campaigns/{campaign_id}/leads`, records at `campaign_leads`) and
  `campaign_replies` (`GET /v1/campaigns/{campaign_id}/replies`, records at `replies`): both use
  `fan_out` (`ids_from.request` against `campaigns`' own `GET /v1/campaigns` endpoint,
  `into.path_var: campaign_id`, `stamp_field: campaign_id`) to drive one sub-sequence per campaign,
  stamping the source campaign's id onto every emitted record — there is no way to discover a lead
  or reply's association with a campaign except by first listing campaigns and querying each one's
  sub-resources.

Every new list endpoint's response envelope is uniform: `{has_more, next_page, status, errors,
<resource_key>: [...]}` — confirmed directly from the live spec's per-endpoint response schema.

## Write actions & risks

**Pass B addition** (no legacy counterpart; legacy `persistiq.Write` always returned
`ErrUnsupportedOperation`).

- `update_lead` (`PATCH /v1/leads/{{ record.id }}`, `path_fields: ["id"]`, `body_type: json`):
  updates `status`/`status_id`/`owner_id`/`bounced`/`optedout`/`data` on an existing lead. **This
  corrects a prior version of this bundle**, which declared `PUT` with a guessed flat field set
  including a nonexistent `title` field — the live spec confirms the real method is `PATCH` and the
  real flat top-level fields are exactly those 6 (`data` itself is a nested free-form object of
  lead data fields to update, not further expanded here). Changing `status`/`status_id`/`owner_id`
  can move a lead into or out of active outbound-sequence automation; approval required.
- `create_campaign` (`POST /v1/campaigns`, `body_type: json`, requires `campaign_name`+`owner_id`):
  creates a new campaign. Approval required.
- `duplicate_campaign` (`POST /v1/campaigns/duplicate`, `body_type: json`, requires
  `campaign_id`+`owner_id`, optional `name`): duplicates an existing campaign's steps/sequence into
  a new one. Approval required.
- `add_lead_to_campaign` (`POST /v1/campaigns/{{ record.campaign_id }}/leads`, `path_fields:
  ["campaign_id"]`, `body_type: json`): enrolls a lead into a campaign (`lead_id`/`mailbox_id`/
  `skip_if_exists`/`override_lead_limit`/`leads`, none required per the live spec). The lead may
  start receiving automated outreach immediately depending on campaign schedule/state. Approval
  required.
- `remove_lead_from_campaign` (`DELETE /v1/campaigns/{{ record.campaign_id }}/leads/{{ record.id
  }}`, `path_fields: ["campaign_id", "id"]`, `body_type: none`, `delete.missing_ok_status: [404]`):
  removes a lead from a campaign, stopping further scheduled outreach in that sequence. Approval
  required.
- `reply_to_campaign_message` (`POST /v1/campaigns/{{ record.campaign_id }}/replies`,
  `path_fields: ["campaign_id"]`, `body_type: json`, requires `inbox_message_id`+`body`): sends a
  real outbound email reply on behalf of the campaign's mailbox owner. Irreversible once delivered.
  Approval required.
- `add_dnc_domain` (`POST /v1/dnc_domains`, `body_type: json`, requires `name`): blocks future
  outreach to a domain account-wide. Approval required.

`create_lead` (`POST /v1/leads`) is **not implemented** — see Known limits (`ENGINE_GAP`).
`PUT /v1/webhook_plugin` (repoints webhook-notification URLs) is **not implemented** — deferred for
the same webhook-URL-mutation caution as every other webhook write in this migration.

## Known limits

- **`ENGINE_GAP`: `create_lead` (`POST /v1/leads`) is not dialect-expressible.** The live OpenAPI
  spec confirms its request body requires a top-level `lead` wrapper object
  (`{"lead": {...fields...}, "creator_id": "...", "dup": "update"|"skip"}`) — an envelope shape.
  The engine's declarative `WriteAction` dialect always sends either the record's own fields
  directly (default JSON body) or an explicit `body_fields` allow-list of top-level record keys —
  there is no mechanism to wrap the outgoing body in a named object-valued envelope key. Recorded in
  `api_surface.json` as `out_of_scope` with this exact reasoning.
- **This bundle's prior review incorrectly assessed `persistiq.com/api-docs` as a dead/parked
  placeholder page.** A direct fetch of `https://persistiq.com/api-docs/v1/swagger.json` during
  THIS review returned a live, current, 14-path OpenAPI 3.0.1 spec — the docs are reachable. This
  review's entire surface/write research supersedes the prior pass's community-client-only
  grounding with this live spec as primary source; the community Node.js client
  (`github.com/LFCENG/persistiq`) and Singer tap (`github.com/NickLeoMartin/tap-persistiq`) remain
  useful corroborating secondary sources but are no longer the sole source of truth.
- **`/v1/mailboxes`, `/v1/activities`, and `/v1/accounts` — 3 of the 6 legacy-parity streams — do
  NOT appear anywhere in the live 14-path spec.** This could mean they are undocumented-but-still-
  live endpoints, deprecated, or were never a real top-level resource at all (the live spec's own
  `user_type` schema nests a `mailboxes[]` array INSIDE each user, suggesting `/v1/mailboxes` may
  never have been its own top-level list resource). Per this migration's rule that legacy is ground
  truth over documentation when the two conflict, these 3 streams are kept EXACTLY as legacy
  implements them (same path, same records key, same parity tests) — not removed, not "corrected"
  against a spec that may itself be incomplete. This is an honest documented limitation, not a
  silent behavior change; a future pass with live credentials should verify whether these 3
  endpoints still respond before either keeping or formally deprecating them.
- **Real pagination envelope carries `has_more`/`next_page`, not reconciled with legacy's
  page-number scheme.** Every live list-response schema includes `has_more` (boolean) and
  `next_page` (nullable string) fields alongside the resource array, suggesting the real API may
  support (or have migrated to) cursor-style pagination. Legacy's own `page_number` pagination
  (`page`/`per_page` query params) is independently confirmed as real and accepted (both are listed
  as query parameters on `GET /v1/leads`), so this bundle keeps that exact, already-proven-working
  shape for every stream rather than switching to an unverified `next_page`-based scheme without
  live-credential confirmation of its exact semantics (absolute URL? page token? bare page number?
  the schema types it as a bare nullable string with no further detail). Fixture pages use the
  page-number-based `page`/`per_page` request shape throughout, consistent with legacy.
- Legacy accepts runtime `page_size` and `max_pages` config values, but the declarative engine only
  supports fixed bundle-authored pagination integers. This bundle honors the legacy default page
  size and intentionally does not declare ignored `page_size`/`max_pages` `spec.json` properties.
- Fixtures use synthetic values only; every new stream's fixture shape mirrors the live spec's own
  documented response schema (field names, nesting, the uniform `has_more`/`next_page`/`status`/
  `errors`/`<resource>` envelope) rather than a guess.
