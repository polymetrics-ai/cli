# Overview

Bitly is a declarative-HTTP bundle reading and writing the Bitly v4 REST API
(`https://api-ssl.bitly.com/v4/...`). It originated as a wave1-pilot migration (PLAN.md P-3) of
`internal/connectors/bitly` (the legacy hand-written connector, which stayed read-only and covered
only `organizations`/`groups`/`campaigns`/`bitlinks`); this Pass B pass expands the bundle to the
API's full practical surface — see `api_surface.json` for the endpoint-by-endpoint accounting. The
legacy package remains registered and unchanged until wave6's registry flip; every NEW stream/write
added in this pass has no legacy counterpart to stay parity-tested against (`capabilities.write`
was `false` in wave1; `docs/migration/conventions.md` §5's parity-deviation ledger therefore has no
new entries here — there is no legacy write behavior to diverge from).

## Auth setup

Provide a Bitly OAuth access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api-ssl.bitly.com/v4` and may be overridden for tests/proxies.

## Streams notes

Ten read streams, all full-refresh (Bitly's core resource endpoints expose no incremental cursor
field on either side of this pass):

- **`organizations`**, **`groups`**, **`campaigns`** — simple non-paginated list endpoints (`GET
  /organizations`, `/groups`, `/campaigns`); records live at the top-level key matching the stream
  name.
- **`channels`** — `GET /channels`; records at `channels`. Channels group bitlinks under a
  campaign for link-in-bio-style presentation.
- **`bsds`** (branded short domains) — `GET /bsds` returns a single JSON object whose `bsds` key is
  a bare array of domain strings (`{"bsds": ["bit.ly", ...]}`), not an array of objects — there is
  no per-domain object shape to explode into multiple records (`connsdk.RecordsAt` only expands an
  array of JSON *objects*; a bare string array yields zero records, and `records.keyed_object` only
  explodes an object's *values*, which must themselves be objects — see the parity-deviation
  ledger's ip2whois `nameservers` entry, `docs/migration/conventions.md` §5 item 12, for the same
  documented `ENGINE_GAP` this dialect still has for scalar-array fan-out). This stream is
  therefore modeled as a **single-record snapshot**: `records.path: ""` selects the whole response
  body as one record, and a static-literal `computed_fields` marker (`"account": "self"`) supplies
  the schema's primary key since the API itself has no natural per-row id here. One row per sync,
  containing the full domain list as an array field — an honest, lossless representation, not a
  workaround.
- **`webhooks`** — `GET /webhooks`; records at `webhooks`. The full webhook object schema
  (`guid`/`group_guid`/`campaign_guid`/`url`/`event`/`is_active`/`client_id`/`created`/`modified`/
  `updated_by`) is reconstructed from Bitly's aggregate public documentation footprint (the create/
  update tutorial page and the webhook event-payload shape both partially describe it; the
  reference sub-page for this endpoint did not fully render during this pass's fetch) rather than
  a directly observed live response — flagged here per Known Limits below.
- **`qr_codes`** — `GET /qr-codes`; records at `qr_codes`.
- **`group_tags`** — `GET /groups/{group_guid}/tags`, scoped by the same `config.group_guid` the
  `bitlinks` stream uses. Same bare-string-array shape as `bsds` (`{"tags": ["a","b"]}`), so it
  uses the identical `records.path: ""` single-record pattern; `computed_fields` stamps
  `group_guid` from config onto the one emitted record (there is no response-body field naming
  which group the tags belong to).
- **`bitlinks`** — unchanged from wave1: group-scoped (`/groups/{{ config.group_guid }}/bitlinks`),
  `next_url` pagination (`pagination.next`, Bitly's own absolute-URL convention), `size=50` static
  per-stream query, records at `links`.

Single-object **detail** endpoints for an already-covered list stream (`GET
/organizations/{organization_guid}`, `/groups/{group_guid}`, `/campaigns/{campaign_guid}`,
`/channels/{channel_guid}`, `/webhooks/{webhook_guid}`, `/qr-codes/{qr_code_id}`,
`/bitlinks/{bitlink}`) are declared `excluded: {category: "duplicate_of"}` in `api_surface.json` —
each returns the identical record shape already reachable through its list stream's per-item
object, so a separate stream would only add cost without adding coverage.

The click/engagement/scan analytics matrix (~30 endpoints: per-bitlink and per-group breakdowns by
country/city/device/referrer/referring-domain/referring-network, time-series, top-N rankings,
shorten-count and feature/historical-usage aggregates, custom-bitlink click breakdowns) is excluded
`out_of_scope` for this pass as a deliberate breadth-vs-cost triage call, not an oversight — see
`api_surface.json`'s `scope` field and Known Limits.

## Write actions & risks

Fifteen write actions, none present in legacy (legacy shipped `capabilities.write: false`):

- **`create_bitlink`** / **`update_bitlink`** / **`delete_bitlink`** — full bitlink lifecycle.
  `update_bitlink`'s `long_url` change consumes a Bitly encode-limit unit and immediately redirects
  all future traffic on that short link; `delete_bitlink` is irreversible.
- **`update_bitlink_tags`** / **`delete_bitlink_tags`** — bitlink tag-set mutation.
  `update_bitlink_tags` REPLACES the full tag set (not an additive merge) — Bitly's PATCH
  `/bitlinks/{bitlink}/tags` semantics, matched here verbatim.
- **`create_campaign`** / **`update_campaign`** — campaign lifecycle (no delete endpoint exists in
  the API).
- **`update_group`** / **`update_group_preferences`** — group rename/re-parent and default-domain
  preference mutation (no group create/delete endpoint exists in the API; groups are provisioned
  through the Bitly web console only).
- **`create_channel`** / **`update_channel`** — channel lifecycle (no delete endpoint exists in the
  API).
- **`create_webhook`** / **`update_webhook`** / **`delete_webhook`** — webhook subscription
  lifecycle. Creating or updating a webhook's `url` registers/redirects live event delivery (click/
  scan notifications) to an external endpoint of the caller's choosing — review the target before
  enabling, per `metadata.json`'s `risk.write`.
- **`create_custom_bitlink`** / **`update_custom_bitlink`** — custom keyword (branded back-half)
  lifecycle; no delete endpoint exists (custom bitlinks are removed by re-pointing or through the
  web console). `update_custom_bitlink` re-points an already-live custom URL at a different
  bitlink — this redirects real traffic immediately.
- **`create_qr_code`** / **`update_qr_code`** / **`delete_qr_code`** — QR code resource lifecycle.
  `update_qr_code`'s `destination` change redirects anyone scanning an already-printed/distributed
  code; `delete_qr_code` is irreversible for any distributed copy.

Every action's per-record `risk` string in `writes.json` is the authoritative, reviewable summary;
`metadata.json`'s `risk.write`/`risk.approval` roll these up for the connector as a whole.

## Known limits

- **The click/engagement/scan analytics matrix is not modeled** (see Streams notes and
  `api_surface.json`). Each of these ~30 endpoints returns a narrow dimensional cut (by country,
  city, device, referrer, time bucket, or ranking) of the same underlying click/scan event data;
  implementing all of them as separate streams was judged disproportionate breadth-for-cost for
  this pass and none has a legacy counterpart to preserve parity against. A future pass MAY add a
  small representative subset (e.g. a single `bitlink_clicks_summary` stream) if a concrete use
  case needs it; none was added here because the single-object summary shape needs a per-bitlink
  config coupling (`config.bitlink`) for what is otherwise a one-row-per-sync result, which did not
  clear this pass's practicality bar (see the excluded `/bitlinks/{bitlink}/clicks/summary` entry's
  reason).
- **`bsds` and `group_tags` are single-record snapshots, not one-record-per-item streams** — see
  Streams notes above for the full `RecordsAt`/`keyed_object` reasoning. This is the same
  `ENGINE_GAP` class documented in `docs/migration/conventions.md` §5 item 12 (ip2whois's
  `nameservers`); it is not re-opened as a new gap here, just re-encountered and worked around the
  same honest way (a single wrapping record, not a fabricated per-element fan-out).
- **The `webhooks` schema's field set is reconstructed from Bitly's aggregate documentation**, not
  a directly observed live response (see Streams notes) — treat `campaign_guid`/`client_id`/
  `updated_by` as best-effort until validated against a real account's webhook list.
- **No list endpoint exists for `custom_bitlinks`** — `GET /custom_bitlinks/{custom_bitlink}` needs
  an id already known from elsewhere (e.g. the `custom_bitlink` value returned by
  `create_custom_bitlink`'s own response), so it is excluded `requires_elevated_scope` rather than
  modeled as a stream; `create_custom_bitlink`/`update_custom_bitlink` writes are still fully
  supported.
- **`page_size`/`max_pages` are not runtime-configurable** on `bitlinks` (carried from wave1,
  unchanged): the engine's `next_url` paginator has no config-driven page-size knob, so `size=50`
  remains a static per-stream query literal and neither `page_size` nor `max_pages` is declared in
  `spec.json`.
- **Legacy's fixture-mode-only fields are not modeled** (carried from wave1, unchanged): legacy's
  `connector`/`fixture`/`previous_cursor` fixture-mode markers are not part of the live record
  shape and are out of scope for this bundle's schemas.
