# Overview

Pabbly Subscriptions Billing started as a read-only declarative migration of
`internal/connectors/pabbly-subscriptions-billing` (legacy Go connector) and has been expanded to
full documented API-surface coverage in Pass B. It reads customers, subscriptions, plans,
invoices, and products, and writes customer/subscription/product/plan/coupon/payment-method/addon/
addon-category/license mutations through the real Pabbly Subscriptions Billing REST API
(`https://payments.pabbly.com/api/v1`). Legacy only ever implemented the 4 read streams; every
write action and the `products` stream are new Pass-B capability, not a legacy port — legacy stays
registered and unchanged until wave6's registry flip.

## API surface (Pass B)

The bundle's `docs_url` (`https://www.pabbly.com/subscriptions/api/`) now redirects to
`apidocs.pabbly.com`, a marketing shell with no endpoint content behind it ("Failed to load
products" — confirmed live). The real, complete endpoint list (base URL
`https://payments.pabbly.com/api/v1`, Basic auth with an `apikey`/`secretkey` pair) was recovered
from a Wayback Machine capture of the same docs page
(`https://web.archive.org/web/20211128152246/https://www.pabbly.com/subscriptions/api/`) — see
`api_surface.json` for the full 40-endpoint breakdown. Every endpoint is now `covered_by` a
stream/write action or `excluded` with a specific reason; there are no blanket "Pass B" exclusions
left. Most exclusions are `duplicate_of` (single-resource-by-id GETs with no id-enumeration source,
already superseded by their list-endpoint counterpart) or `requires_elevated_scope` (per-parent
sub-resource lists — coupons, payment methods, checkout pages, transactions, refunds, custom
fields, addons/addon-categories by product — that would need a `fan_out` over a parent id list this
bundle does not yet model); `verifyhosted`/`portal_sessions` are `out_of_scope` (customer-facing
token flows, not data mutations); the 3 reporting/stats endpoints
(`getdashboardstats`/`revenuetransaction`/`mrrsubscription`) are `non_data_endpoint` (aggregate
snapshots with no stable per-record identity).

## Auth setup

Provide a Pabbly account `username` (config) and `password` (secret); they are sent via HTTP Basic
auth (`auth: [{"mode": "basic", "username": "{{ config.username }}", "password": "{{
secrets.password }}"}]`) exactly like legacy's `connsdk.Basic(username, password)`. Legacy
hard-errors when either value is unset (`pabbly-subscriptions-billing connector requires config
username and secret password`), matching this bundle's `required: ["username", "password"]`. The
real Pabbly API documents this same Basic-auth pair as `apikey`/`secretkey`; this bundle keeps
legacy's `username`/`password` spec field names for config-surface stability rather than renaming
them (same credential shape, no behavior change).

## Streams notes

All 4 legacy-parity streams (`customers`, `subscriptions`, `plans`, `invoices`) share the identical
shape: `GET` against the Pabbly list endpoint, records at `data`, primary key `["id"]`. `customers`,
`subscriptions`, and `invoices` declare `x-cursor-field: created_at` on their schemas (matching
legacy's catalog-only `CursorFields` declaration on those 3 streams); `plans` declares none
(legacy's `plans` stream carries no `CursorFields` either). No stream declares a `streams.json`
`incremental` block: legacy's `Read`/`harvest` never applies a server-side or client-side
incremental filter for any of the 4 streams — the `CursorFields` catalog metadata is descriptive
only, never wired into an actual request parameter — so this bundle matches that real behavior
exactly rather than inventing an incremental filter legacy never had.

Pagination for the 4 legacy-parity streams follows legacy's own `next_page`-in-body convention: the
response body's `next_page` field carries the literal value of the NEXT `page` query parameter to
send — modeled as `pagination.type: cursor` with `token_path: next_page` and `cursor_param: page`.
Pagination stops when `next_page` is absent or empty, identical to legacy's
`strings.TrimSpace(next) == ""` check; no `stop_path` is declared since legacy has no separate
boolean stop signal beyond the token itself. `per_page` is sent on every request from the
`page_size` config value (default `100`, matching legacy's `defaultPageSize`) via each stream's
`query.per_page` object-form entry (`default: "100"`).

`products` (new Pass-B stream, `GET /products`) is a `passthrough`-projected, single-page stream
(`pagination.type: none`) matching the real documented response shape (`{"status", "message",
"data": [...]}`, one page of every product with no pagination parameters documented for this
endpoint at all).

## Write actions & risks

Pass B adds 13 write actions on top of the pre-existing 7 (`create_customer`/`update_customer`/
`cancel_subscription`/`create_product`/`update_product`/`create_plan`/`update_plan`):
`create_subscription` (assigns a plan to an existing customer and starts recurring billing),
`update_subscription` (changes an existing subscription's plan/payment terms),
`create_coupon` (creates a discount coupon scoped to a product),
`create_payment_method`/`update_payment_method` (stores/replaces a customer's card on file via the
connected payment gateway), `create_addon`/`update_addon`/`delete_addon` (sellable plan add-ons),
`create_addon_category`/`update_addon_category`/`delete_addon_category` (add-on organizing
categories), and `create_license`/`update_license` (license-key pools for a product's plan).
`delete_addon` and `delete_addon_category` are `kind: "delete"` with `confirm: "destructive"`,
matching `cancel_subscription`'s existing destructive-confirm pattern. `metadata.json` now declares
`capabilities.write: true` (previously incorrectly left `false` even though `writes.json` already
carried the original 7 actions — a stale leftover from before this bundle's write surface was
added, fixed as part of this pass).

## Known limits

- `page_size`/`max_pages` config validation (legacy's numeric-range and `all`/`unlimited` keyword
  parsing) is not reproduced at the bundle-config level; the engine treats `page_size` as an opaque
  string substituted directly into the `per_page` query param. This never changes emitted record
  DATA for any legacy-valid input; it only narrows client-side input validation.
- No `incremental` block is declared on any stream, matching legacy's real (lack of) incremental
  filtering behavior exactly — not a narrowing. `x-cursor-field` remains declared on 3 of the 4
  legacy-parity schemas purely as catalog/candidate-cursor metadata, mirroring legacy's own
  `CursorFields` field.
- **`base_url` still points at the legacy-parity host (`https://www.pabbly.com/subscriptions/api`),
  not the real documented API host (`https://payments.pabbly.com/api/v1`).** The 4 pre-existing
  streams' paths (`/customers`, `/subscriptions`, `/plans`, `/invoices`) were themselves modeled
  against the legacy Go connector's own base URL and path shape, not the real Pabbly docs (which
  were unreachable at migration time and only recovered via the Wayback Machine for this Pass B
  research pass). Migrating the base_url and all paths to the real host is a bigger, riskier change
  than this capability-expansion pass should make silently — it would require re-verifying every
  existing stream's pagination/record-envelope assumptions against the real API rather than
  legacy's, which may not match (e.g. real `GET /customers` returns `{"status","message","data":
  [...]}` with no visible pagination fields in the documented example, unlike legacy's assumed
  `next_page` cursor). New Pass-B write actions and the `products` stream use the REAL documented
  paths/host implicitly through the same `{{ config.base_url }}` templating, so a caller who
  overrides `base_url` to the real host gets correct write behavior; the default `base_url` is
  unchanged to avoid breaking existing legacy-parity read behavior. Reconciling the two hosts is
  flagged for a follow-up increment, not silently patched over here.
- The `requires_elevated_scope`-excluded per-parent sub-resource lists (coupons, payment methods,
  checkout pages, transactions, refunds, custom fields, per-product addons/addon-categories) are a
  real, documented gap: this dialect's `fan_out` mechanism could close them (iterating the
  `customers`/`products` streams' ids), but that is additional stream work beyond this pass's
  write-action focus and is tracked in `api_surface.json` rather than silently modeled.
