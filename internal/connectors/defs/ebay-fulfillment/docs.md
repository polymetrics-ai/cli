# Overview

eBay Fulfillment is a Tier-2 (hooks) declarative-HTTP migration. It reads a seller's orders,
exploded line items, a shipment-oriented projection, and payment disputes through the eBay Sell
Fulfillment REST API (`GET {{ config.base_url }}/sell/fulfillment/v1/...`). This bundle targets
full capability parity with `internal/connectors/ebay-fulfillment` (the hand-written connector it
replaces); the legacy package stays registered and unchanged until wave6's registry flip.
Read-only (`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide an eBay OAuth 2.0 refresh token via the `refresh_token` secret, plus the OAuth client
credentials (`username` = client id, `password` = client secret, both secrets). `base.auth`
declares `mode: custom, hook: ebay-fulfillment` — the engine has no built-in
`oauth2_refresh_token` auth mode (only `oauth2_client_credentials`, a different grant), so
`internal/connectors/hooks/ebay-fulfillment/hooks.go`'s `AuthHook` exchanges the refresh token for
a short-lived access token at `refresh_token_endpoint` (default
`https://api.ebay.com/identity/v1/oauth2/token`), authenticating the token request with HTTP
Basic `client_id:client_secret` and a `grant_type=refresh_token` form body, caching the resulting
access token until 60s before its declared expiry — porting legacy's `refreshTokenAuth`
(`ebay-fulfillment/auth.go`) verbatim. `scope` is optional and omitted from the token-request form
entirely when unset, matching legacy. The hook fails closed on a non-https or hostless
`token_url` (stricter than legacy's http-tolerant `validateURL`; documented below).

**All dynamic (fixture-replay) conformance checks are bundle-level `skip_dynamic`'d** (`metadata.json`),
matching gmail's identical shape: conformance's synthetic config overrides `token_url` with the
non-URL literal `"synthetic-conformance-value"` (`runtimeConfigForEngine` ignores `spec.json`
defaults), which the `AuthHook`'s `validateHTTPSURL` correctly rejects — every auth-resolving
dynamic check would otherwise fail identically on this synthetic-config artifact, not a real
bundle defect. The `fixtures/streams/{orders,payment_disputes}/{page_1,page_2}.json` and
`fixtures/streams/{order_line_items,shipping_fulfillments}/page_1.json` files are still committed
in full recorded-real-shape (§4 fixture rules) — they simply aren't exercised by `conformance`
today; they become live the moment a `paritytest/ebay-fulfillment` suite or a real access token
drives `engine.Read`/`engine.Check` directly (bypassing the synthetic-config auth wall), which is
recommended follow-up work, not part of this pass.

## Streams notes

`orders` and `payment_disputes` are fully declarative:

- `orders` — `GET /sell/fulfillment/v1/order` (records at `orders`), `offset_limit` pagination
  (`limit`/`offset` query params). `computed_fields` flattens `buyer.username` ->
  `buyer_username` and `pricingSummary.total.{value,currency}` -> `total_value`/`total_currency`;
  `line_item_count` is derived with the `length` filter over the raw `lineItems` array;
  every other field is a bare `{{ record.<camelCase> }}` rename (typed extraction preserves the
  raw wire type). Primary key `order_id`; incremental cursor `creation_date`, sent as an eBay
  `creationdate:[<value>..]` `filter` query param (`omit_when_absent: true` — absent on a
  full-refresh read with no persisted cursor and no `start_date` configured).
- `payment_disputes` — `GET /sell/fulfillment/v1/payment_dispute` (records at
  `paymentDisputeSummaries`), also `offset_limit` pagination, `amount.value`/`amount.currency`
  flattened via `computed_fields`. **The `filter` query param is the SAME `creationdate:[...]`
  literal used by `orders`, even though this stream's own timestamp/cursor field is
  `openDate`/`open_date`, not `creationDate`.** This reproduces legacy's `dateFilter` exactly:
  legacy's `Read` computes `filter := dateFilter(req)` ONCE, generically, from whatever cursor the
  CURRENT stream persists, and always wraps it in the literal `creationdate:[...]` range-filter
  syntax regardless of which stream is being read — an existing legacy accepted-input quirk this
  migration reproduces, not "corrects" (matching `docs/migration/conventions.md`'s meta-rule and
  ebay-finance's identical documented `transfers`/`transactionDate` quirk).

`order_line_items` and `shipping_fulfillments` are `StreamHook`-handled
(`internal/connectors/hooks/ebay-fulfillment/hooks.go`), re-projecting the SAME `GET
/sell/fulfillment/v1/order` page (records at `orders`) two different ways, exactly like legacy's
`emitProjected` switch:

- `order_line_items` — explodes each order's `lineItems[]` array into one record per line item,
  carrying the parent order's `order_id`/`creation_date`. This is a nested-array-per-parent-record
  fan-out the declarative dialect cannot express: `records.path` selects exactly one array
  location per page, and `records.keyed_object` explodes a JSON OBJECT's values, not array
  elements nested inside a parent record — there is no declarative primitive for "one record
  becomes N records, each carrying fields from both the parent and a nested array element."
- `shipping_fulfillments` — projects each order's
  `fulfillmentStartInstructions[0].shippingStep.shipTo.{fullName,contactAddress.*}` into a
  shipment-oriented row. `computed_fields`' `record.<path>` resolution walks `map[string]any`
  only (`internal/connectors/engine/interpolate.go`'s `resolveRecordPathValue`); it has no
  array-index reference, so accessing `fulfillmentStartInstructions[0]` cannot be expressed as a
  `computed_fields` template.

Both hook streams follow eBay's absolute `next` URL when present, or advance `offset` on a full
page with no `next` — matching legacy's `harvest` exactly (the hook reimplements this loop
directly against `rt.Requester`, the engine's already-built Requester: base URL and the
refresh-token Bearer auth are already resolved declaratively by the `AuthHook`; the `StreamHook`
adds only the fetch/pagination/explosion logic, never touches auth). The SAME
`dateFilter`-equivalent helper (keyed off the stream's own persisted incremental cursor, falling
back to `start_date` config) builds the `creationdate:[...]` filter for both hook streams too.

Primary keys: `orders`/`shipping_fulfillments` on `order_id`; `order_line_items` on
`line_item_id`; `payment_disputes` on `payment_dispute_id`. `x-cursor-field` on each schema names
the stream's own timestamp field (`creation_date`/`open_date`) even where the query filter's field
name differs, above.

## Write actions & risks

None. Legacy `ebay_fulfillment.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`orders`/`payment_disputes` use `offset_limit` pagination rather than legacy's `next`-URL-first
  (falling back to offset) scheme.** Legacy's `harvest` follows eBay's optional absolute `next`
  URL when the API supplies one, falling back to `offset` advancement only when `next` is absent.
  The engine's declarative `next_url` pagination type stops entirely when the body's `next` path
  is absent — it has no offset fallback — and a genuine 2-page `next_url` fixture is not statically
  authorable (the next URL is the replay server's own dynamically-assigned address, unknown until
  runtime; `docs/migration/conventions.md` §4's sanctioned single-page exception requires a live
  `paritytest` companion, out of scope for this bundle). `offset_limit` pagination models the
  offset-fallback path — the common case, and the only one with static, predictable request query
  values — reproducing every field for any input legacy's offset path would accept; the optional
  `next`-URL-follow branch (used only when eBay's API happens to return one) is not modeled for
  these two declarative streams. **`order_line_items`/`shipping_fulfillments`'s `StreamHook` DOES
  model the full `next`-URL-first-then-offset-fallback behavior** (ported verbatim from legacy's
  `harvest`), since a hook can reuse the replay server's real request/response cycle without
  needing a statically-authored `next` URL value.
- **`page_size` fixture-vs-live separation.** `orders`/`payment_disputes` declare a stream-level
  `pagination.page_size: 2` (not legacy's real default, 50) purely so their required 2-page
  conformance fixtures (`docs/migration/conventions.md` §4) stay small and readable — matching
  ebay-finance's identical documented `transactions` deviation. This never changes which records
  are emitted for an in-range request, only request cadence/count; it also means `config.page_size`
  has no effect on these two streams (the paginator's fixed `page_size` wins outright — declaring a
  redundant `query.limit` template would be immediately overridden by the paginator's own query
  every page, per `engine/read.go`'s `mergeQuery`, so it is not declared at all). The
  `order_line_items`/`shipping_fulfillments` `StreamHook`, by contrast, DOES read
  `config.page_size` directly (1-1000, default 50, matching legacy's `resolvePageSize` bounds) — a
  malformed/unset value falls back to the default rather than erroring.
- **`api_host`/`base_url` config-surface narrowing.** Legacy accepts either `api_host` or
  `base_url` (whichever is set; `api_host` wins if both are), plus a Basic-credential
  `refresh_token_endpoint` override, each independently scheme/host-validated
  (`ebayBaseURL`/`tokenEndpoint`/`validateURL`). This bundle exposes a single `base_url` config
  key (default `https://api.ebay.com`) and `refresh_token_endpoint` (default
  `https://api.ebay.com/identity/v1/oauth2/token`); the `AuthHook` validates
  `refresh_token_endpoint` itself (fails closed on non-https/hostless, intentionally stricter than
  legacy's http-or-https tolerance — never wrong for eBay's real, always-https token endpoint,
  documented parity deviation). `base_url`'s own scheme/host is not independently re-validated by
  this bundle (the engine has no declarative URL-shape validator for `base.url` the way legacy's Go
  code did); a malformed `base_url` surfaces as a generic request-construction/connection error
  instead of legacy's specific message.
- **`max_pages`** is not modeled (F6, dead config): legacy's `resolveMaxPages` accepts an optional
  cap, but nothing in this bundle (declarative or hook) has a config-driven per-request-count
  override mechanism for it, so it is not declared in `spec.json`.
- **`order_line_items`'s `total_value`/`total_currency`** are only populated when the raw line
  item's own nested `total` object is present — matches legacy's `lineItemRecord` exactly (no
  fallback to the parent order's `pricingSummary.total`).
- The full eBay Sell Fulfillment API surface (single-order lookup, shipping-fulfillment
  sub-resource CRUD, payment-dispute accept/contest) is out of scope for this wave; see
  `api_surface.json`'s `excluded` entries.
