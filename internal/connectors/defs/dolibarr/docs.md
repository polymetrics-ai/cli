# Overview

Dolibarr was quarantined in wave1 for an `ENGINE_GAP`: Dolibarr's REST API is genuinely 0-indexed
(legacy sends `page=0` for the first page, `page=1` for the second, per `dolibarr.go`'s `for page :=
0; ...` harvest loop), and the engine's `page_number` pagination could not express a 0-indexed
start (a plain Go `int` `StartPage` field could not distinguish an explicit `0` from an omitted
key). This gap was closed by the S4 engine mini-wave's `PaginationSpec.StartPage *int`
(`"start_page": 0` is now distinguishable and honored verbatim) — this bundle is the unblock build
using that dialect addition. It reads Dolibarr third parties, contacts, products, customer
invoices, and orders (list + single-record detail) through the Dolibarr REST API, and writes
create/update/delete/validate mutations for the same 5 objects. This bundle migrates
`internal/connectors/dolibarr` (the hand-written connector it replaces at capability parity), then
extends it in Pass B with 5 detail streams and 17 write actions covering every CRUD/validate REST
method on Dolibarr's `api_thirdparties`/`api_contacts`/`api_products`/`api_invoices`/`api_orders`
PHP classes (fetched directly from the Dolibarr/dolibarr GitHub repository's `develop` branch); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Dolibarr API key via the `api_key` secret; it is sent as the `DOLAPIKEY` request header
(`api_key_header` auth mode) and is never logged.

## Streams notes

All 5 streams (`thirdparties`, `contacts`, `products`, `invoices`, `orders`) share the identical
shape: `GET` against the Dolibarr list endpoint, records at the response body root (`records.path:
""`, matching legacy's top-level-JSON-array `connsdk.RecordsAt(resp.Body, "")`), and every request
sends `sortfield=t.rowid&sortorder=ASC` as static query params (matching legacy's `harvest` query
construction exactly). Pagination is genuinely 0-indexed (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `start_page: 0`, `page_size: 100` matching legacy's
`dolibarrDefaultPageSize`) — the first request sends `page=0`, matching legacy's loop exactly; a
page returning fewer than `limit` records stops the read (legacy's `len(records) < pageSize`
short-page stop).

Each of the 5 list streams declares a bare `incremental.cursor_field: date_modification`, matching
legacy's published `CursorFields` and preserving incremental sync-mode eligibility. No
`request_param` is declared because legacy's `harvest` sends no server-side filter derived from a
cursor or `start_date`-shaped config value at all (only the static
`sortfield`/`sortorder`/`limit`/`page` params above). `x-cursor-field: date_modification` is also
declared on every schema so the cursor field remains part of the projected record shape.

`thirdparty_detail`/`contact_detail`/`product_detail`/`invoice_detail`/`order_detail` (Pass B
additions) each read `GET /{resource}/{id}` — a single-object detail record scoped by a new
`{resource}_id` config field (e.g. `thirdparty_id`). Every detail schema widens its list-stream
counterpart with a few genuinely detail-only fields the Dolibarr object classes publish (e.g.
`address`/`code_client`/`siren`/`siret`/`tva_intra` for thirdparties, `description`/`weight` for
products) — these are never present on the list endpoint's page-array entries. `pagination:
{"type": "none"}` overrides the base 0-indexed `page_number` pagination for all 5, matching the
single-object-stream pattern used elsewhere in this codebase (see dockerhub's `namespace`/
`repository_detail`).

## Write actions & risks

17 write actions cover every create/update/delete/validate REST method Dolibarr's 5 already-migrated
object classes expose (per their real `@url`-annotated PHP methods):

- `create_thirdparty`/`update_thirdparty`/`delete_thirdparty` (`POST`/`PUT`/`DELETE /thirdparties`)
- `create_contact`/`update_contact`/`delete_contact` (`POST`/`PUT`/`DELETE /contacts`)
- `create_product`/`update_product`/`delete_product` (`POST`/`PUT`/`DELETE /products`)
- `create_invoice`/`update_invoice`/`delete_invoice`/`validate_invoice` (`POST`/`PUT`/
  `DELETE /invoices`, `POST /invoices/{id}/validate`)
- `create_order`/`update_order`/`delete_order`/`validate_order` (`POST`/`PUT`/`DELETE /orders`,
  `POST /orders/{id}/validate`)

`validate_invoice`/`validate_order` are `kind: update` state-transition writes (not `create`/
`delete`): Dolibarr's own `validate()` PHP method accepts an optional JSON body
(`{idwarehouse, notrigger}`, both defaulting server-side) alongside the path `id` — modeled with
`body_fields: ["idwarehouse", "notrigger"]` so the body is restricted to exactly those two optional
fields rather than the default "every field except path_fields" behavior. Validating is a real,
one-way state transition (draft to validated, assigning the invoice/order's final reference
number) — every validate action's `risk` field calls this out explicitly. `update_invoice`/
`update_order`/`delete_invoice`/`delete_order` are only accepted by Dolibarr while the record is
still in draft status (Dolibarr itself enforces this business rule server-side; this bundle does
not attempt to pre-check status client-side, matching how every other write action in this
codebase defers business-rule enforcement to the live API).

## Known limits

- **Dolibarr's 404-as-empty-page end-of-data signal is approximated, not reproduced 1:1.** Legacy
  treats an HTTP 404 ("No record found") past the end of the data set as a clean end-of-data signal
  identical to a short/empty 200 page (`isNotFound(err)` inside `harvest`, and again in `Check`).
  The engine's `page_number` paginator's only stop signal is a short/empty successful page; it does
  not special-case a 404 response as "clean end of data" — a 404 propagates as a request error
  instead. In practice this only diverges from legacy when the LAST page happens to be exactly a
  multiple of `page_size` (so the very next page is empty AND some deployments return 404 rather
  than `200 []` for it); every deployment/page shape observed in fixtures and the parity-relevant
  legacy tests returns a `200 []`-style empty/short page, which both sides handle identically. Not
  modeled as an engine change since it recurs only for this one connector (below the §6 recurrence
  threshold) and true production Dolibarr instances vary in whether they emit 404 vs. empty-200 for
  an out-of-range page.
- **Base URL is `base_url`-only; the legacy `my_dolibarr_domain_url` bare-domain convenience is
  dropped.** Legacy accepts either an explicit `base_url` override OR a bare
  `my_dolibarr_domain_url` (e.g. `"mydomain.com/dolibarr"`), deriving
  `https://<domain>/api/index.php` from the latter in Go (`domainToBaseURL`). The engine's
  `spec.json` `"default"` materialization only fills in a FIXED literal default, not a
  config-value-derived one (conventions.md §3's `spec.json "default"` paragraph — "for a DERIVED
  default ... this mechanism alone is not enough; either require `base_url` and drop the derivation
  ... or express the derivation as a `computed_fields`-style template if/when the dialect grows
  one for base-URL construction"); no such mechanism exists yet. This bundle requires the full
  `base_url` (e.g. `https://your-dolibarr-host/api/index.php`) and does not declare
  `my_dolibarr_domain_url` at all (a declared-but-unwireable key is worse than an absent one, per
  conventions.md F6). Documented config-surface narrowing, not a silent behavior change for any
  input this bundle itself accepts.
- Full Dolibarr API surface beyond the 5 already-migrated objects (users, projects, warehouses/
  stock, bank accounts, expense reports, proposals/quotes, shipments, contracts) remains out of
  scope; see `api_surface.json`'s `excluded` entries. Within the 5 migrated objects, every
  sub-resource family (categories/tags, notifications, bank accounts, purchase prices, variant
  attributes, line items, payments, contact-linking, recurring-invoice templates, discount/
  credit-note operations) is also out of scope this pass — each is individually reasoned in
  `api_surface.json`, mostly `out_of_scope` (niche sub-object, no legacy precedent) or
  `requires_elevated_scope` (financial-instrument or identity-provisioning data this bundle
  deliberately does not touch).
- Only the forward `validate` state transition is modeled for invoices/orders; the reverse
  `settodraft`/`reopen` rollback transitions, and the `settopaid`/`settounpaid`/`setinvoiced`
  status-shortcut endpoints (which change financial state without a genuine underlying payment/
  invoice record), are excluded as `out_of_scope`/`requires_elevated_scope` respectively — see
  `api_surface.json`.
