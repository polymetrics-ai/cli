# Overview

CallRail reads and writes CallRail call tracking, contact, and account-configuration data through
the CallRail v3 REST API (`https://api.callrail.com/v3/...`). This bundle targets capability parity
with `internal/connectors/callrail` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

**Pass B full-surface expansion** (`api_surface.json`, reviewed 2026-07-04 against the live
apidocs.callrail.com docs site): every documented resource category is now accounted for. Beyond
the original 4 legacy streams (`calls`, `companies`, `users`, `text_messages`), this bundle adds 13
more: `accounts`, `tags`, `trackers`, `form_submissions`, `integrations`, `integration_filters`,
`notifications`, `caller_ids`, `sms_threads`, `message_flows`, `leads`, and 2 fan-out sub-resource
streams (`page_views` per-call, `lead_timeline` per-lead) — 17 streams total. `capabilities.write`
is now `true`, with 27 write actions covering tags, companies, users, calls (metadata +
outbound-call placement), text messages, integrations, integration filters, notifications, caller
ids, message flows, SMS threads, and tracker reconfiguration. See `api_surface.json` for the full
endpoint-by-endpoint accounting, including every deliberate exclusion and its real,
closed-vocabulary reason.

## Auth setup

Provide a CallRail API key via the `api_key` secret; it is sent as `Authorization: Token
token="<api_key>"` (with the value itself quoted inside the header, matching legacy's
`fmt.Sprintf("Token token=%q", apiKey)` at `callrail.go:254`) via `auth.mode: api_key_header` with
`prefix: "Token token="` and a `value` template (`"\"{{ secrets.api_key }}\""`) that supplies the
surrounding literal quotes; the secret itself is never logged. The required `account_id` config
value is substituted into every request path (`/a/{{ config.account_id }}/...`, urlencoded by
`InterpolatePath`'s per-segment default, matching legacy's own `url.PathEscape(account)` in
`accountPath`). `base_url` defaults to `https://api.callrail.com/v3` and may be overridden for
tests/proxies.

## Streams notes

All 4 streams (`calls`, `companies`, `users`, `text_messages`) share the same pagination shape:
CallRail's page-number convention (`pagination.type: page_number`, `page_param: page`,
`size_param: per_page`) — a page shorter than `page_size` stops pagination, which is functionally
equivalent to legacy's own primary stop signal (`total_pages` reached) for every real CallRail
response (the last page is never longer than `per_page`); the one edge case where a result set is
an exact multiple of `per_page` costs one extra, empty-page request on the engine side that legacy's
`total <= page` check would have avoided, never a data difference (both sides stop after emitting
the same records). `page_size` is `100`, matching legacy's own default (see Known limits for why it
is not runtime-configurable).

Each of the 4 legacy streams sends `start_date` (`param_format: date`, converting a resolved RFC3339
lower bound to `YYYY-MM-DD` for the wire) computed either from the sync's persisted cursor or, on a
fresh sync, from the `start_date` config value — matching legacy's `startDateParam`/`dateOnly`
exactly for the RFC3339-input case (see Known limits for the accepted-input narrowing this
requires). Per-stream cursor fields match legacy's own `CursorFields` declarations: `calls` ->
`start_time`, `companies`/`users` -> `created_at`, `text_messages` -> `last_message_at`.

**New Pass B streams**: `accounts` (`GET /a.json`, account-scoped path segment not needed), `tags`,
`trackers`, `notifications`, `sms_threads`, `caller_ids` are unpaginated-incrementally-useful lists
sharing the base page_number pagination; `tags`/`trackers` additionally declare
`incremental.cursor_field: created_at` off a `start_date`-driven state cursor (client-tracked, since
CallRail's tags/trackers list endpoints document no server-side date filter param — the engine has
no incremental `request_param` for these two, matching how legacy never implemented them either).
`form_submissions` reuses the same `start_date`/`param_format: date` request-param shape as the
4 legacy streams, cursor field `submitted_at`. `integrations`, `integration_filters`, `caller_ids`,
`message_flows`, and `leads` support an optional `company_id` query filter (`stream.Query`'s
`omit_when_absent` dialect, config key `company_id`) matching each endpoint's documented optional
company-scoping parameter — omitted entirely when unset. `trackers`' nested `company: {id, name}`
object is flattened via `computed_fields` into `company_id`/`company_name`;
`integration_filters`' nested `integration: {id, type}` is flattened into `integration_type`
the same way.

`page_views` and `lead_timeline` are **fan-out sub-resource streams** (`stream.fan_out`, matching
campayn's established pattern): `page_views` first lists every call id (`ids_from.request` against
`/calls.json`), then issues `GET /calls/{{ fanout.id }}/page_views.json` per call id, stamping the
parent `call_id` onto every emitted page-view record (page views have no `id` field of their own,
so the schema's primary key is the composite `[call_id, created_at]`). `lead_timeline` follows the
identical shape off `/leads.json` -> `GET /leads/{{ fanout.id }}/timeline.json`; the endpoint returns
a single JSON object (not an array) at the `lead` path, so `records.path: "lead"` naturally yields
exactly one record per lead id via `connsdk.RecordsAt`'s bare-object-is-one-record behavior.

## Write actions & risks

`capabilities.write` is now `true` (Pass B). 27 actions, grouped by resource:

- **Tags**: `create_tag`, `update_tag` (low risk), `delete_tag` (irreversible — removes the tag from
  every call/text it was ever applied to; approval recommended).
- **Companies**: `create_company`, `update_company` (setting `status: disabled` deactivates all of a
  company's tracking numbers — approval recommended for status changes specifically). CallRail has
  no true company DELETE (the docs call it "Disabling a Company"), so `update_company` is the only
  covered mutation; see `api_surface.json`.
- **Users**: `create_user`/`update_user` (administrator-scoped API key required by CallRail itself;
  approval recommended), `delete_user` (irreversible, approval required).
- **Calls**: `update_call` (tags/note/lead_status/value/customer_name metadata; this is also how
  CallRail applies tags to a call — there is no separate tagging endpoint), `create_outbound_call`
  (places a REAL phone call between two numbers; US/Canada only; approval required — a genuine
  real-world side effect, not just a CallRail data mutation).
- **Text messages**: `send_text_message` (sends a REAL SMS/MMS; approval required; the `media_url`
  variant is covered, direct multipart file-upload MMS is not — see Known limits).
- **Integrations**: `create_integration`/`update_integration` (Webhooks/Custom types only, matching
  what the API itself allows creating), `disable_integration` (the docs' own term; not a true
  delete).
- **Integration filters**: `create_integration_filter`, `update_integration_filter`,
  `delete_integration_filter` (all low risk — these narrow which calls trigger an existing
  integration, they don't touch the integration itself).
- **Notifications**: `create_notification`, `update_notification`, `delete_notification` (all low
  risk, user-alert-subscription config).
- **Outbound caller IDs**: `create_caller_id` (triggers a REAL verification phone call to the
  registered number; approval required), `delete_caller_id` (low risk).
- **SMS threads**: `update_sms_thread` (notes/value/tags/lead_qualification metadata; low risk).
- **Trackers**: `update_tracker` only (reconfigures an already-provisioned tracker's call flow,
  whisper message, SMS setting, or source rules) — tracker CREATE and the docs' own "Disabling a
  Tracker" DELETE both provision/deprovision a real, billable phone number and are excluded as
  `requires_elevated_scope`/`destructive_admin` respectively; see `api_surface.json`.
- **Message flows**: `create_message_flow`, `update_message_flow` (the docs' own PUT endpoint takes
  no `{message_flow_id}` path segment — the flow is identified purely by the body's `id` field, so
  this action declares no `path_fields`), `delete_message_flow`.

## Known limits

- **`start_date` config input is narrowed to RFC3339 (or bare Unix-seconds), no longer bare
  `YYYY-MM-DD`.** Legacy's `startDateParam` (`callrail.go:266-282`) accepts EITHER a bare
  `YYYY-MM-DD` value or full RFC3339 for the `start_date` config value, narrowing either to a date
  string itself before sending it as the `start_date` query param. The engine's `param_format: date`
  conversion (`formatParam`/`parseLowerBoundTime`) only accepts an all-digits (Unix-seconds) value or
  a full RFC3339 timestamp — a bare `"2026-01-01"` fails to parse as RFC3339 and hard-errors. This
  bundle's `spec.json` therefore declares `start_date` as `format: date-time` (RFC3339 only), a
  documented config-surface narrowing versus legacy's more permissive YYYY-MM-DD-or-RFC3339
  acceptance; any RFC3339 `start_date` value (e.g. `"2026-01-01T00:00:00Z"`) still produces the exact
  same `YYYY-MM-DD` wire value legacy would send for the equivalent date.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config overrides
  (`callrailPageSize`/`callrailMaxPages`, `callrail.go:348-376`, `page_size` defaulting to 100,
  capped at 250). The engine's `page_number` paginator's `PageSize` is a static bundle-authored int
  (not templated), and there is no `MaxPages`-equivalent config-driven knob either; `max_pages` is
  unbounded (matching legacy's own `max_pages=0`/`all`/`unlimited` default). `page_size` is fixed at
  `100` to match legacy's own default exactly; the conformance fixture for `calls` is a single page
  of 3 records (all `total_records`) — a short page relative to `page_size: 100` — so
  `pagination_terminates` observes exactly one request, matching the real one-request-in-production
  behavior for any result set under 100 records; `companies`/`users`/`text_messages` are likewise
  single fixture pages.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a `previous_cursor` field onto every fixture-mode record
  when a prior cursor happens to be set (`callrail.go:206-239`); this is not part of the live record
  shape. This bundle's schemas and fixtures target the live path only.
- **`send_text_message` does not cover multipart-form MMS file upload.** The docs show 3 ways to
  send an MMS: a publicly-hosted `media_url` (covered), or a multipart `-F media_file=@path` upload
  (not covered — the engine's write dialect has no multipart/binary-payload body type; this would
  be a `binary_payload`-category exclusion if enumerated as its own endpoint, but since it is a body
  variant of the same `POST /text-messages.json` endpoint already covered by `send_text_message`,
  it is documented here rather than as a separate `api_surface.json` line).
- **`tags`/`trackers` incremental cursoring is client-tracked, not server-filtered.** Neither
  endpoint documents a server-side date-range filter query parameter (unlike `calls`/
  `form_submissions`, which send `start_date`), so `incremental.cursor_field: created_at` on these
  2 streams relies purely on the engine's persisted-cursor comparison against each stream's own
  `created_at` field; every page is still fetched and filtered client-side implicitly by the cursor
  advancing, not by a narrower request. This is not a deviation — legacy never implemented these two
  streams at all, so there is no prior behavior to diverge from.
- Every remaining known CallRail endpoint is either covered or excluded with a specific reason; see
  `api_surface.json`'s `excluded` entries for the full, closed-vocabulary accounting (analytics
  aggregates, phone-number-provisioning actions, and admin-config surfaces judged out of scope for
  this pass).
