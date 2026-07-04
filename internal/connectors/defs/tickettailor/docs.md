# Overview

Ticket Tailor is a wave2 fan-out declarative-HTTP migration, since Pass B expanded to the full
documented Ticket Tailor REST API v1 surface (published OpenAPI spec at
`https://app.tickettailor-stitching.com/openapi.yml`, linked from
`https://developers.tickettailor.com/`). It reads and writes events, orders, issued tickets, event
series, holds, discounts, membership types, issued memberships, products, stores, vouchers,
checkout forms/elements, and a box-office overview (`https://api.tickettailor.com/v1/<resource>`).
This bundle is capability-parity migrated from `internal/connectors/tickettailor` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip. `capabilities.write` is now `true`.

**Important discovery — pagination was fixed, not just extended (see `api_surface.json`'s `scope`
note for the full account):** the real Ticket Tailor API's actual pagination convention for every
list endpoint — including the 3 original streams — is cursor-based (`starting_after`/
`ending_before`/`limit`, a `data[]`/`links{next,previous}` envelope), not the `page`/`limit`
page-number convention the legacy connector and the original wave2 bundle both used. `page` is not
a documented query parameter on any Ticket Tailor list endpoint; sending it would have been
silently ignored, and the real cursor param (`starting_after`) was never sent at all, meaning the
original bundle would never have advanced past page 1 against the live API. This was a genuine
functional defect carried over from the legacy Go connector (`tickettailor.go:87`, itself never
verified against production), not a deliberately documented parity decision — nothing in this
bundle's prior `docs.md` called it out as an intentional deviation. It is fixed in this pass for
all 3 original streams plus every new stream.

## Auth setup

Provide a Ticket Tailor API key via the `api_key` secret; it is sent as the username of HTTP
Basic auth with an empty password (`Authorization: Basic base64(<api_key>:)`), matching legacy's
`connsdk.Basic(key, "")` (`tickettailor.go:107`) and the published OpenAPI spec's own `BasicAuth`
security scheme (`"Use base64 encoded api key created in Ticket Tailor dashboard"`) exactly, and is
never logged. `base_url` defaults to `https://api.tickettailor.com/v1`, matching legacy's
`defaultBaseURL` fallback and the spec's own default server.

## Streams notes

All 17 streams share `next_url` pagination (`base.pagination`: `type: next_url`, `next_url_path:
links.next`, `allow_cross_host: true`) reading the real, documented `links.next` field — a
same-host RELATIVE path+query string per Ticket Tailor's own documented example (e.g.
`/v1/events?starting_after=ev_123`), not an absolute URL. `allow_cross_host: true` is required
purely because a host-less relative value fails the engine's origin-check guard exactly like a
genuinely cross-host URL would (`checkOrigin`'s `u.Host == ""` branch); there is no actual
cross-host risk here since the value is always same-origin by construction. `base.query`'s static
`limit` param sends `100` on the FIRST request of each stream's sequence; subsequent pages replay
whatever query the server's own `links.next` value encodes (Ticket Tailor's own pagination
contract, not something this bundle controls). Per conventions.md's sanctioned `next_url`
single-page-fixture exception (a static fixture cannot embed the correct absolute replay-server URL
for a second page), every stream ships exactly one fixture page; `pagination_terminates` exercises
`events` (the bundle's first declared stream) — a single fixture page is sufficient to prove one
consumed request terminates cleanly, matching the sanctioned pattern's own reasoning.

**Legacy-parity streams (field shape unchanged, only pagination fixed):** `events`, `orders`,
`issued_tickets` — same `GET /<resource>`, `data` envelope, `"projection": "passthrough"` (legacy's
`connsdk.Harvest` callback emits every raw field verbatim, `tickettailor.go:88` — no `mapRecord`-
style filtering step; default `"schema"` projection mode would silently drop every real Ticket
Tailor field not named in each stream's declared schema properties, an undocumented silent
data-shape change relative to legacy's raw passthrough), primary key `["id"]`.

**New Pass B streams (14):**

- `event_series`, `holds`, `discounts`, `membership_types`, `issued_memberships`, `products`,
  `stores`, `vouchers`, `checkout_forms` — plain top-level `GET /<resource>` list streams, same
  `data` envelope and `next_url` pagination shape, `"projection": "passthrough"` (matching the
  original 3 streams' own reasoning: the live API returns considerably more per-object detail than
  any hand-picked field list would capture), primary key `["id"]`.
- `voucher_codes` (`GET /vouchers/{voucher_id}/codes`), `checkout_form_elements` (`GET
  /checkout_forms/{checkout_form_id}/elements`), `event_series_overrides` (`GET
  /event_series/{event_series_id}/overrides`), `event_series_waitlist_signups` (`GET
  /event_series/{event_series_id}/waitlist_signups`) — all `fan_out` streams (`ids_from.request`
  over their respective parent list endpoint, `into: {path_var: ...}`, `stamp_field: ...`) since
  each is only reachable per-parent-id with no top-level list endpoint of its own — the sanctioned
  dialect mechanism for exactly this shape (conventions.md §3 "Sub-resource fan-out"). None of the 4
  has a globally-unique per-record id on its own (`voucher_codes`' own `id` is only unique within
  one voucher's codes, etc.); `[<parent>_id, id]` is each one's genuine composite primary key,
  matching this migration's ip2whois `[domain, role]` composite-key precedent (conventions.md
  ledger item 11).
- `overview` (`GET /overview`) — a single-object box-office statistics stream
  (`records: {"path": ".", "single_object": true}`, `pagination.type: none`). The documented
  `Overview` schema has no natural per-record identifier at all (it is a single aggregate, not a
  list); `computed_fields: {"id": "overview"}` stamps the sanctioned static-literal constant as the
  primary key (a template with no `{{ }}` markers passes through verbatim, conventions.md §3), the
  same mechanism searxng's `stream` marker field uses.

## Write actions & risks

25 write actions, `body_type: "form"` throughout (matching the published spec's own
`application/x-www-form-urlencoded` request bodies) unless noted:

- **Event series**: `create_event_series`, `update_event_series` against
  `/event_series[/{id}]`, `delete_event_series` (destructive/confirm-gated — permanently deletes
  every occurrence within it), and `change_event_series_status` (POST `.../status`,
  `body_fields: ["status"]` — setting `draft`/`sales_closed` immediately stops public ticket sales).
- **Discounts**: `create_discount`, `update_discount`, `delete_discount` against
  `/discounts[/{id}]`.
- **Holds**: `delete_hold` only (DELETE `/holds/{id}`, path-only, no body) — see Known limits for
  why `create_hold`/`update_hold` are NOT implemented despite being documented.
- **Check-ins & tickets**: `create_check_in` (POST `/check_ins`), `create_issued_ticket` (POST
  `/issued_tickets`, issues a ticket directly bypassing checkout), `void_issued_ticket` (POST
  `.../void`).
- **Orders**: `update_order` (buyer contact/address fields) and
  `confirm_order_payment_received` (marks an offline/manual-payment order as paid).
- **Memberships**: `create_membership_type`, `delete_membership_type` against
  `/membership_types[/{id}]`; `create_issued_membership`, `update_issued_membership`,
  `void_issued_membership` against `/issued_memberships[/{id}][/void]`.
- **Vouchers**: `create_voucher`, `update_voucher`, `delete_voucher` against `/vouchers[/{id}]`,
  and `void_voucher_code` (POST `/vouchers/{voucher_id}/codes/{id}/void`, path-only, no body).
- **Products**: `create_product`, `update_product`, `delete_product` against `/products[/{id}]`.

`metadata.json` now declares `capabilities.write: true`; `risk.approval` names
`delete_event_series` as the sole action requiring approval.

## Known limits

- **`create_hold`/`update_hold` are intentionally NOT implemented (`ENGINE_GAP`, documented in
  `api_surface.json`).** The documented request body's required `ticket_type_id` field is a JSON
  object (`additionalProperties`, e.g. `{tt_1: 1, tt_2: 5}`) sent inside an
  `application/x-www-form-urlencoded` body. This dialect's form-body construction
  (`engine/write.go`'s `buildForm`/`stringifyAny`) can only stringify a nested object value as a
  single JSON-string-valued form field (`ticket_type_id=%7B%22tt_1%22%3A1%7D`) — it has no
  mechanism to emit the bracket-notation (`ticket_type_id[tt_1]=1`) or repeated-key encoding a
  nested object inside a form body typically requires, and there is no way to verify which encoding
  Ticket Tailor's real endpoint actually expects without a live test. Guessing would silently risk
  sending a request the real API rejects or misinterprets, so this is documented as a genuine gap
  rather than shipped unverified; `delete_hold` (path-only, no body) is unaffected and implemented
  normally.
- **Pagination was fixed for the 3 original streams, a behavior change from the prior bundle
  revision** (see Overview's "Important discovery" note and `api_surface.json`'s `scope` note for
  the full account): `page_number` (`page`/`limit`) never matched Ticket Tailor's real documented
  API at all, so no caller could have been genuinely relying on the old shape working beyond a
  single page against the live service — the fix is a correctness repair, not a parity-breaking
  scope choice, and no `parity_deviations[]` entry applies since there was no working prior
  behavior for a real caller to diverge from.
- Deep event-series sub-resource CONFIGURATION mutations (ticket types, ticket groups, bundles,
  schedule overrides, single-occurrence create/update/delete, single checkout-form-element update,
  single-store update) are excluded as Pass B breadth-vs-cost triage, matching bitly's precedent for
  narrower nested-configuration sub-resources — see `api_surface.json`'s `out_of_scope` exclusions
  for the complete, endpoint-by-endpoint list and reasoning.
- `check_ins` (the raw check-in event log, as opposed to the `create_check_in` write) and
  `issued_membership_redemptions` are not modeled as read streams/writes respectively — see
  `api_surface.json`'s exclusions.
- No incremental cursor is modeled on any stream (matching the original bundle's own behavior):
  every sync is full-refresh. The published spec documents rich `created_at`/`updated_at`
  greater-than/less-than filter parameters on several endpoints that could drive a genuine
  server-side incremental filter in a future increment; this bundle does not wire them (a scope
  choice, not a blocker).
