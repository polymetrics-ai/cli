# Overview

Eventzilla is a fan-out declarative-HTTP migration. It reads Eventzilla events, categories,
users, attendees, ticket types, and transactions through the Eventzilla v2 REST API (`GET
https://www.eventzillaapi.net/api/v2/...`). This bundle covers the 5 streams
`internal/connectors/eventzilla` (the hand-written connector) implements, plus a new
`transactions` stream and 2 write actions (`checkin_attendee`, `toggle_event_sales`) added in the
Pass B full-surface expansion against the real Eventzilla v2 developer docs
(https://developer.eventzilla.net/docs/); the legacy package stays registered and unchanged until
wave6's registry flip. `attendees` and `tickets` — previously blocked (`ENGINE_GAP`; see
`docs/migration/status.json`'s `partial[]` entry) — are expressed via the engine's `fan_out`
dialect (S4 engine mini-wave item 2); `transactions` reuses the identical fan_out shape. Eventzilla
now has 2 writes; `capabilities.write` is `true`.

## Auth setup

Provide the Eventzilla API key via the `api_key` secret; it is sent as the `x-api-key` request
header (`{"mode":"api_key_header","header":"x-api-key","value":"{{ secrets.api_key }}"}`), matching
legacy's `connsdk.APIKeyHeader(eventzillaAPIKeyHeader, secret, "")` exactly (`eventzilla.go`'s
`requester`, `eventzillaAPIKeyHeader = "x-api-key"`). `base_url` defaults to
`https://www.eventzillaapi.net/api/v2` and may be overridden for tests/proxies.

## Streams notes

`events`, `categories`, and `users` are top-level list endpoints (`GET /events`, `/categories`,
`/users`); records live at the resource-named JSON key. Eventzilla returns `{"<field>":[...]}`
with no total/has_more marker, so pagination stops on a short page
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`), matching legacy's
`harvest` function exactly (`len(records) < pageSize` stop condition). `page_size` is declared as
a fixed literal (`100`) in `base.pagination`, matching legacy's own default
(`eventzillaDefaultPageSize = 100` in `eventzilla.go`, also the config-bounds cap) — the engine's
`offset_limit` paginator has no config-driven page-size override mechanism (see Known limits), so
this bundle bakes in legacy's default rather than legacy's config-driven override range.

`categories`' primary key is `["category"]` (the category name string itself), matching legacy's
own `PrimaryKey: []string{"category"}` — Eventzilla's category list has no separate id field.

`attendees` and `tickets` are per-event sub-resource fan-out reads, matching legacy's
`readSubstream` (`eventzilla.go:141-160`) exactly: `fan_out.ids_from.request` issues a paginated
`GET /events` sequence (reusing the `attendees`/`tickets` stream's own effective pagination — the
base `offset_limit` block, since neither stream declares its own override), extracting `id_field:
id` off every discovered event record; `into.path_var: "event_id"` threads each discovered event id
into `/events/{{ fanout.id }}/attendees` (or `/tickets`). The child streams deliberately do NOT
declare `stamp_field`: Eventzilla's live attendee/ticket payload already carries `event_id` as a
JSON integer, and legacy preserves that raw value whenever it is present (`stampParent` only filled
the field for the legacy fallback case where the raw child record omitted it). Leaving the raw
`event_id` untouched keeps emitted data and type fidelity with legacy for the real API shape.

`transactions` (Pass B addition) uses the same event-id fan-out source as `attendees`/`tickets`:
`GET /events/{event_id}/transactions`, records at the `transactions` field path, primary key
`checkout_id` (the numeric id both single-transaction detail GETs — by `checkout_id` or by
`refno` — key off of; both are excluded from `api_surface.json` as `duplicate_of` since they
return the identical record shape this stream already covers). The real Eventzilla docs do not
show the exact pagination envelope for `/events/{event_id}/transactions`, so this bundle infers the
same `{"transactions":[...]}` resource-named-key envelope every other Eventzilla list endpoint
uses (`events`/`categories`/`users`/`attendees`/`tickets` all share this shape) and reuses the base
`offset_limit` pagination block — a reasonable, documented inference, not a verified-live shape;
flagged in Known limits.

## Write actions & risks

Eventzilla now supports 2 write actions, added in the Pass B full-surface expansion:

- **`checkin_attendee`** (`POST /attendees/checkin`) — marks an attendee checked in (or reverts
  check-in) by scanning their ticket barcode (`barcode`, `eventcheckin`). Low-risk operational
  door-scan mutation; no approval required.
- **`toggle_event_sales`** (`POST /events/togglesales`) — publishes or unpublishes an event's
  public sales page (`eventid`, `status`). Setting `status: false` immediately stops new ticket
  sales for that event; approval required before use.

Both are single-request JSON-body mutations with no compound follow-up requests, so both are fully
Tier-1 declarative (`writes.json`, no hook needed). The multi-step registration/checkout flow
(`checkout/prepare` -> `checkout/create` -> `checkout/fillorder` -> `checkout/confirm`) and the
order `confirm`/`cancel` endpoints are excluded from this bundle (`api_surface.json`,
`out_of_scope`/`destructive_admin`) — they are compound, stateful, multi-request workflows that
charge/refund real money and cannot be expressed as a single declarative write action.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, capped at 100) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`eventzillaPageSize`/`eventzillaMaxPages` in `eventzilla.go`). The engine's `offset_limit`
  paginator has no config-driven page-size or max-pages knob (`PaginationSpec.PageSize` is a fixed
  literal read once at bundle-load time, not a per-request template), so this bundle declares a
  fixed `page_size: 100` in `base.pagination` (legacy's own default/cap value,
  `eventzillaDefaultPageSize`) and does not declare `page_size`/`max_pages` in `spec.json` at all (a
  declared-but-unwireable config key is worse than an absent one, per conventions.md F6
  precedent). Pagination is bounded only by Eventzilla's own short-page stop signal, matching
  Eventzilla's real termination behavior.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a broad synthetic superset of fields
  shared across all 5 streams' mappers (including `attendees`/`tickets`-only fields), which is not
  the live wire shape any single migrated stream actually returns. This bundle's schemas and
  fixtures target the LIVE record shape only (`eventzilla.go`'s `harvest`/`mapRecord` functions),
  per the bitly-pilot precedent (`docs/migration/conventions.md`'s worked example): the engine's
  own fixture-replay conformance harness supersedes the need for an in-connector fixture-mode
  branch.
- **`transactions`' pagination envelope is inferred, not verified live (Pass B addition).** The
  public Eventzilla developer docs document `GET /events/{event_id}/transactions`'s response
  fields per transaction but do not show an explicit pagination wrapper or an `offset`/`limit`
  query-param contract for this specific endpoint. This bundle infers the same
  `{"transactions":[...]}` resource-named-key envelope and `offset_limit` pagination every other
  Eventzilla list endpoint in this bundle uses (a consistent pattern across `events`/`categories`/
  `users`/`attendees`/`tickets`), since Eventzilla's own docs describe a single uniform pagination
  convention (`offset`/`limit` query params) applied API-wide. If the real endpoint's envelope
  differs, this is a documented inference risk, not a silently-guessed shape.
- **`checkin_attendee`/`toggle_event_sales` are single-request writes with no response-shape
  validation beyond `write_request_shape`'s method/path/body match.** Neither action's real
  response body is consumed by any follow-up logic (no `WriteHook` needed), so this bundle does
  not assert anything about the response payload beyond what `fixtures/writes/*.json`'s optional
  `response` block documents for replay.
