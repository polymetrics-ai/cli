# Overview

Veeqo is an inventory and order management platform. This bundle reads orders, products,
customers, warehouses, suppliers, purchase orders, sales channels ("stores"), delivery methods,
and tags from the Veeqo API (`{base_url}`, default `https://api.veeqo.com`), and writes
create/update/delete mutations for orders, products, customers, suppliers, warehouses, delivery
methods, tags, sales channels, product properties, payments, and shipments. It originally migrated
`internal/connectors/veeqo` (190 loc, a read-only single-stream connector), which stays
registered and unchanged until wave6's registry flip; this Pass B pass expands the bundle using
Veeqo's full documented API surface (the published OpenAPI 3.0 document at
`developers.veeqo.com/api-schemas/veeqo-api/` plus every individual operation doc page reachable
from the site's sitemap, reviewed 2026-07-04).

## Auth setup

Veeqo authenticates via a single secret, `api_key`, sent as the `x-api-key` header with an empty
prefix, matching legacy's `connsdk.APIKeyHeader("x-api-key", apiKey, "")` requester exactly
(`veeqo.go:120`). Expressed via the engine's `api_key_header` auth mode.

## Streams notes

`orders` is the original legacy-parity stream: `GET /orders`, records from the response body root
(`"records": {"path": "."}`), no pagination, `id` force-cast to a string via `last_path_segment`,
and the optional `start_date` config-only query passthrough (all preserved unchanged from the
original migration — see the original rationale for why this is a stateless passthrough, not a
true incremental cursor).

Every other stream is new in this Pass B pass. Veeqo's documented pagination is `page`/`page_size`
query parameters (confirmed directly from the live OpenAPI document's declared `parameters` on
every list operation), expressed via the engine's `page_number` pagination type
(`page_param: page`, `size_param: page_size`, `start_page: 1`, `page_size: 100`); every new stream's
`records.path` is `"."` (root array), matching Veeqo's own convention of returning a bare JSON
array at every list endpoint (confirmed from the OpenAPI document's response examples for
`/products`, `/customers`, `/warehouses`, `/suppliers`, `/channels`, `/delivery_methods`):

- `products` (`GET /products`) — every field the raw API returns is large and deeply nested
  (variants, images, inventory); this bundle projects only `id`/`title`/`notes`/`created_at`/
  `updated_at` per the schema-as-projection rule, matching this connector's practical read surface
  rather than the full nested catalog object.
- `customers` (`GET /customers`), `warehouses` (`GET /warehouses`), `suppliers`
  (`GET /suppliers`), `purchase_orders` (`GET /purchase_orders`), `channels` (`GET /channels`,
  Veeqo's "stores"), `delivery_methods` (`GET /delivery_methods`).
- `tags` (`GET /tags`) — no pagination declared: Veeqo's OpenAPI document lists no `page`/
  `page_size` parameters on this endpoint (tag lists are typically small, account-wide).

Every new stream's `id` is force-cast to a string via `last_path_segment` for consistency with the
`orders` stream's existing convention, since Veeqo's real wire shape for every resource's `id` is a
JSON integer.

No stream declares an `incremental` block for the new streams: while `products` documents
`created_at_min`/`updated_at_min` query filters and `orders` documents `since_id`, this bundle
does not wire a persisted-cursor `incremental` block for any NEW stream in this pass (kept
narrowly scoped to the same full-refresh shape `orders` already has) — a genuine opportunity for
a future increment, not modeled here to avoid scope creep beyond the practical breadth-vs-cost
target for this pass.

## Write actions & risks

24 write actions, all `body_type: json`, covering every resource whose Veeqo create/update body
this dialect can express — including nested-envelope bodies (`{"customer": {...}}`,
`{"order": {...line_items_attributes...}}`, `{"product": {...product_variants_attributes...}}`,
`{"shipment": {...tracking_number_attributes...}}`) and partial envelopes
(payments' `{"amount": ..., "payment_attributes": {...}}`): `body_type: json` sends every
top-level `record_schema` key verbatim, and any one of those keys can itself be a nested object —
the same pattern `yotpo`/`teamtailor`/`teamwork`'s envelope-shaped writes already use elsewhere in
this repo. An earlier pass of this bundle incorrectly excluded these as an `ENGINE_GAP`
("the dialect cannot express a nested envelope"); that claim was factually wrong and has been
corrected here.

- **orders**: `create_order`, `update_order`, `cancel_order` (a body-less `PUT
  /orders/{id}/cancel` state transition, `kind: custom`, matching the `braintree`
  `cancel_subscription` pattern for a zero-body action write).
- **products**: `create_product`, `update_product`, `delete_product` (idempotent, 404 tolerated).
- **customers**: `create_customer`, `update_customer`.
- **payments**: `create_payment` (`{"amount": ..., "payment_attributes": {"order_id": ..., ...}}`).
- **shipments**: `create_shipment` (`{"shipment": {"tracking_number_attributes": {...}}, "carrier_id": ..., "allocation_id": ..., "order_id": ...}`).
- **suppliers**: `create_supplier`, `update_supplier`, `delete_supplier` (idempotent, 404
  tolerated).
- **warehouses**: `create_warehouse`, `update_warehouse` (Veeqo documents no
  `DELETE /warehouses/{id}` endpoint).
- **delivery_methods**: `create_delivery_method`, `update_delivery_method`,
  `delete_delivery_method`.
- **tags**: `create_tag`, `delete_tag` (Veeqo documents no `PUT /tags/{tag_id}` single-tag-update
  endpoint — only bulk tag-rename via `PUT /tags`, out of scope, see Known limits).
- **channels**: `create_channel`, `update_channel`, `delete_channel`.
- **product_properties**: `create_product_property` only (a single `{"name": "..."}` body; Veeqo
  documents no update/delete endpoint for this resource).

All actions carry `"risk": "external mutation; approval required"` (destructive deletes add
`"confirm": "destructive"`).

## Known limits

- Full API-surface classification lives in `api_surface.json` (70 endpoint entries reviewed
  2026-07-04): 9 read streams, 24 write actions, and every remaining endpoint excluded with a
  specific, real per-endpoint category and reason (no blanket bucket) — `duplicate_of` for
  single-item detail GETs already covered by their list stream; `out_of_scope` for narrow
  per-parent sub-resources with no top-level list endpoint (order returns/allocations/notes,
  product custom-property-values, bundle contents, per-sellable-per-warehouse stock entries, bulk
  tagging); `requires_elevated_scope` for carrier label-purchase/rate-shopping actions needing
  configured carrier credentials beyond this connector's plain API-key auth; `binary_payload` for
  the PDF/PNG label-download endpoint; and `non_data_endpoint` for the authenticated-account
  profile object.
- `PUT /tags` (bulk tag rename/merge) and `POST`/`DELETE /bulk_tagging` (bulk multi-record
  tag-assignment) are real documented endpoints excluded as `out_of_scope`: they operate over an
  arbitrary list of taggable-type/id pairs rather than a single typed record, which this dialect's
  per-record write model does not fit without a hook (Tier-2 escalation not justified for this
  narrow a capability).
- **`id` on every stream is force-cast to a string** via `last_path_segment` (Veeqo's real wire
  shape for every resource id is a JSON integer) — see the original migration's rationale for
  `orders.id`, applied identically to every new stream for consistency.
- **`Check` dials the network; legacy's `Check` never did** — unchanged from the original
  migration, a deliberate fail-loud improvement with zero record-data impact.
- The optional `start_date` filter on `orders` remains a stateless, config-only passthrough (see
  the original migration) — not a true incremental sync; none of the new streams add incremental
  filtering either (see Streams notes).
