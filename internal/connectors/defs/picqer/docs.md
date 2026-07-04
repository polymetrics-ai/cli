# Overview

Picqer is a wave2 fan-out declarative-HTTP migration, expanded to the full documented v1 surface in
Pass B. It reads 30 Picqer resources (products, customers, orders, picklists, warehouses,
suppliers, tags, purchase orders, receipts, returns, return statuses/reasons, backorders, comments,
stock history, users, product/customer/order fields, pricelists, shipping providers, VAT groups,
locations, location types, picking containers, picklist batches, shipments, packagings, packing
stations, webshop orders, and webhooks) through the Picqer REST API
(`GET https://<organization>.picqer.com/api/v1/<resource>`), and writes 36 practical mutations
(customer/supplier/tag CRUD, product update, order lifecycle, purchase-order lifecycle and
receiving, return CRUD, backorder processing, location/location-type/picking-container/packaging
CRUD, picklist-batch and shipment creation, and webhook create/delete). This bundle originated as a
capability-parity port of `internal/connectors/picqer` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Picqer uses HTTP Basic auth with the API key as the username and an empty password
(`connsdk.Basic(key, password)`, `picqer.go:134`). Provide the key via the `api_key` secret
(preferred) or the plain-config `username` fallback; the dual-candidate `auth` list in
`streams.json` reproduces legacy's exact precedence
(`firstNonEmpty(secret(cfg,"api_key"), cfg.Config["username"])`, `picqer.go:129`): the `api_key`
secret candidate is declared first and gated on its own presence (`when`), falling through to the
`username` config candidate when `api_key` is unset. An optional `password` secret is honored on
either branch (Picqer's own convention leaves it blank).

`organization_name` derives the base URL as `https://<organization_name>.picqer.com/api/v1`,
matching legacy's own derivation (`picqer.go:144`). See Known limits for the one narrowing this
bundle makes versus legacy's config surface.

## Streams notes

All 30 streams are simple list endpoints (`GET /<resource>`) with `"projection": "passthrough"` —
the original 6 legacy-parity streams (`products`, `customers`, `orders`, `picklists`, `warehouses`,
`suppliers`) return every raw field unchanged plus a `computed_fields.id` alias
(`idproduct`/`idcustomer`/`idorder`/`idpicklist`/`idwarehouse`/`idsupplier` copied into a uniform
`id` field via typed bare-reference extraction, preserving Picqer's real integer wire type, no
stringification); the 24 Pass-B-added streams (`tags`, `purchaseorders`, `receipts`, `returns`,
`return_statuses`, `return_reasons`, `backorders`, `comments`, `stockhistory`, `users`,
`product_fields`, `customer_fields`, `order_fields`, `pricelists`, `shippingproviders`,
`vatgroups`, `locations`, `location_types`, `picking_containers`, `picklist_batches`, `shipments`,
`packagings`, `packingstations`, `webshoporders`, `hooks`) follow the identical passthrough +
`computed_fields.id` shape, each keyed off that resource's own `id<resource>`-prefixed primary key.
`receipts` (`GET /receipts`) is Picqer's v2 goods-receiving session resource: `idreceipt`,
`idwarehouse`, `version`, `supplier`/`purchaseorder` linkage objects, `receiptid`, `status`
(`processing`/`completed`), `completed_by`, `amount_received`, `completed_at`, `created`, `updated`,
and the nested `products[]` line-item array survive verbatim under passthrough projection.

Pagination is `offset_limit` with `offset_param: offset` and no `limit_param` — matching legacy's
`connsdk.OffsetPaginator{OffsetParam: "offset", PageSize: size}` (`picqer.go:82`), which never
sends a page-size query parameter at all (`LimitParam` is left empty in legacy); the paginator
stops when a page returns fewer than `page_size` (100) records, a purely client-side threshold,
never a server-enforced page size. This is unchanged for every Pass-B-added stream too.

## Write actions & risks

36 write actions across customer/supplier/tag CRUD, product update, order lifecycle
(pause/resume/reopen/cancel), purchase-order lifecycle (create/mark-as-purchased/close/cancel),
receipt lifecycle (create/complete), return CRUD, backorder processing, location/location-type
CRUD, picking-container CRUD, picklist-batch and shipment creation, packaging CRUD, and webhook
create/delete (`create_hook`/`delete_hook`). See `metadata.json`'s `risk.approval` for the exact
low-risk-vs-approval-required split.

`complete_receipt` (`PUT /receipts/{idreceipt}`) sends Picqer's documented completion body
verbatim — `{"status": "completed"}` — via `body_fields: ["status"]`; Picqer applies received stock
quantities in the background and the synchronous response does not reflect the status change yet
(callers must re-read the `receipts` stream to observe `status: "completed"`). `create_receipt`
(`POST /receipts`) only models the `idpurchaseorder`-required shape; Picqer's v2 API also accepts
`idsupplier` in place of `idpurchaseorder` for a supplier-only (not-yet-purchase-order-linked)
receiving session — see Known limits.

`cancel_order`/`cancel_purchaseorder`/`delete_customer`/`delete_tag`/`delete_return`/
`delete_location`/`delete_hook` are `kind: "delete"` with `missing_ok_status: [404]` (idempotent);
`create_hook`/`create_customer`/`create_supplier`/`create_tag`/`create_purchaseorder`/
`create_receipt`/`create_return`/`create_location`/`create_location_type`/
`create_picking_container`/`create_picklist_batch`/`create_packaging` require no approval
(low-risk, non-destructive); every `update_*`/`pause_order`/`resume_order`/`reopen_order`/
`mark_purchaseorder_as_purchased`/`close_purchaseorder`/`complete_receipt`/`process_backorders`
action requires approval per `metadata.json`.

## Known limits

- **Explicit `base_url` override is not modeled; `organization_name` is now required.** Legacy
  accepts either an explicit `base_url` (checked first) or derives
  `https://<organization_name>.picqer.com/api/v1` when `base_url` is unset (`picqer.go:136-145`).
  The engine's `streams.json` `base.url` is a single non-conditional template (unlike `auth`, which
  supports a `when`-gated candidate list) — there is no mechanism to express "prefer this literal
  override, else derive from this other config key" in one field. Per
  `docs/migration/conventions.md`'s guidance for a derived base URL, this bundle requires
  `organization_name` and drops the `base_url`-override path; a caller who previously pointed the
  legacy connector at a fixed `base_url` (e.g. a proxy or non-standard Picqer deployment) cannot do
  so through this bundle. This is a documented, deliberate config-surface narrowing, not a silent
  behavior change for the common case (an organization-name-driven Picqer tenant).
- **Legacy's defensive `out["id"] == nil` guard is not modeled.** Legacy only back-fills `id` from
  the resource-specific key when the raw record has no pre-existing `id` field at all; since
  Picqer's real API never emits a bare `id` field on these resources (only the prefixed
  `id<resource>` keys), this guard is dead code in practice and the engine's unconditional
  `computed_fields.id` copy is capability parity for every real Picqer response.
- **`create_receipt`'s `idsupplier`-in-place-of-`idpurchaseorder` alternative is not modeled.**
  Picqer's v2 receipts API accepts either `idpurchaseorder` (received against an existing purchase
  order) or `idsupplier` (a supplier-only receiving session that Picqer later reconciles into a
  new purchase order for any unmatched products) — a named-field OR, the same shape as stripe's
  `create_customer` ledger item 1. The draft-07 dialect this engine uses has no `anyOf`/`oneOf`, so
  `create_receipt`'s `record_schema` only models the `idpurchaseorder`-required branch; a
  supplier-only receiving session cannot be started through this action. Out of scope for this
  pass, not silently wrong — `idpurchaseorder` is real, general-purpose Picqer behavior for a
  purchase-order-linked receipt, which is the common case.
- **Sub-resource receiving actions are out of scope.** Picqer's full receiving workflow also
  includes `GET /receipts/{idreceipt}/expected-products`, `POST /receipts/{idreceipt}/products`
  (add a received product line, with `automatic`/`purchaseorder_product`/`new` strategy variants),
  and `POST /receipts/{idreceipt}/products/{idreceipt_product}/revert` — see `api_surface.json` for
  each endpoint's specific exclusion reason. These are fine-grained line-item edits of a receipt
  already covered by the `receipts` stream and `create_receipt`/`complete_receipt` actions, not
  independent business objects; `complete_receipt` covers the terminal lifecycle transition a
  reverse-ETL caller actually needs.
