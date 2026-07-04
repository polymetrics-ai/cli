# Overview

Cin7 is a Pass B full-surface declarative-HTTP connector. It reads and writes Cin7 Core (DEAR
Inventory) products, customers, suppliers, sale/purchase order summaries, product families,
product stock availability, and 8 reference/lookup tables (locations, product categories, brands,
carriers, chart of accounts, payment terms, tax rules, units of measure, price tiers) through the
Cin7 Core External API v2 (`https://inventory.dearsystems.com/externalapi/v2/...`). This bundle
targets capability parity with `internal/connectors/cin7` (the hand-written connector it migrates)
for its 5 original read streams, and substantially exceeds it: legacy was read-only and never
implemented incremental filtering; this bundle adds 10 new read streams, 13 write actions, and
`ModifiedSince`/`UpdatedSince` incremental filtering on all 5 original streams, researched directly
against Cin7 Core's published Apiary API Blueprint (see `api_surface.json`). The legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Cin7 Core authenticates every request with two headers. The application key is a secret: provide
it via the `api_key` secret, sent as the `api-auth-applicationkey` header
(`{"mode": "api_key_header", "header": "api-auth-applicationkey", "value": "{{ secrets.api_key }}"}`)
and never logged. The account id is not secret-shaped and is sent directly as a declared
`base.headers` entry, `api-auth-accountid: {{ config.accountid }}`; `accountid` is `required` in
`spec.json`, matching legacy's own hard requirement (`cin7.go:79-84`, `requester`).

## Streams notes

The 5 legacy-parity streams (`products`, `customers`, `suppliers`, `sale_list`, `purchase_list`)
share the identical Cin7 Core envelope: `GET` against a list resource, records at a
resource-specific top-level array key (`Products`/`CustomerList`/`SupplierList`/`SaleList`/
`PurchaseList`), primary key `["id"]`. `products`/`customers`/`suppliers` additionally send
`IncludeDeprecated=true` on every request, matching legacy's `streamEndpoint.params`. Pagination is
`page_number` (`page`/`limit`, `start_page: 1`, `page_size: 100` — matches legacy's
`cin7DefaultPageSize`); the engine's short-page stop (`recordCount < page_size`) is identical to
legacy's own `harvest` stop condition. Every stream's raw uppercase Cin7 field names (`ID`, `Name`,
`SKU`, ...) are renamed to the schema's lowercase, legacy-matching field names via
`computed_fields` bare `{{ record.<Field> }}` references (typed extraction preserves each field's
native JSON type — `price_tier1`/`cost`/`invoice_amount` stay numeric, matching legacy's own
pass-through of Cin7's numeric JSON values).

**Incremental filtering (new, beyond legacy).** Cin7 Core's Apiary docs confirm `products`/
`customers`/`suppliers` accept a `ModifiedSince` query parameter and `sale_list`/`purchase_list`
accept `UpdatedSince` (both ISO 8601/RFC3339, the engine's default `param_format`) to filter to
records changed since a given timestamp. Legacy's own `harvest` never applied one (a full paginated
sweep every read); this bundle adds `incremental.request_param` to all 5 streams and
`x-cursor-field: last_modified` to their schemas — a genuine new capability, not a parity
requirement, since it never changes what a full-refresh (no prior state) read returns.

**New streams (beyond legacy).** `product_families` (`GET /productFamily` → `ProductFamilies`,
same shape/incremental convention as `products`) covers Cin7's product-family (variant/option)
catalog resource, distinct from `products` (Cin7 models a family of variants — e.g. a T-shirt in
several sizes/colors — as one ProductFamily record referencing N child Product records).
`product_availability` (`GET /ref/productavailability` → `ProductAvailabilityList`) is Cin7's
per-location stock-level resource (`on_hand`/`allocated`/`available`/`on_order`/`stock_on_hand`/
`in_transit`), primary key `["id", "location", "bin"]` (a product's stock is broken out per
warehouse location and, optionally, per bin — the compound key is the documented uniqueness
granularity for this resource); it publishes no incremental filter (undocumented for this
endpoint). 8 reference/lookup-table streams — `locations`, `product_categories`, `brands`,
`carriers`, `chart_of_accounts`, `payment_terms`, `tax_rules`, `units_of_measure` — are small,
`"projection": "passthrough"` GET-list resources (Cin7's own configuration/reference catalogs that
products/customers/suppliers/sales/purchases reference by name or code); `price_tiers` is Cin7's
fixed 10-slot price-tier catalog, the one resource in this bundle whose real response carries no
`Total`/`Page` pagination envelope at all (bare `{"PriceTiers": [...]}`) — it still inherits
`streams.json`'s base `page_number` pagination spec (harmlessly sending unused `page`/`limit` query
params), and terminates normally on its first (and only) short page.

## Write actions & risks

13 write actions across 6 resource groups, all `body_type: "json"` (Cin7 Core's own JSON body
convention for every documented write): **products** (`create_product`/`update_product`, both
`POST`/`PUT /product` — Cin7's own `ID`-in-body-not-in-path PUT convention, so neither action
declares `path_fields`), **customers** (`create_customer`/`update_customer`, `POST`/`PUT
/customer`), **suppliers** (`create_supplier`/`update_supplier`, `POST`/`PUT /supplier`),
**product categories** (`create_product_category`/`update_product_category`/
`delete_product_category` — the only resource group this pass also covers `DELETE` for, since Cin7
Core's own docs show it as a low-risk, narrowly-scoped single-field catalog value; `delete` uses
`path: "/ref/category?ID={{ record.ID }}"` since Cin7's delete convention is a query-string `ID`
param rather than a path segment — the engine's `InterpolatePath` resolves `{{ }}` markers inside a
literal `?`-prefixed query string exactly like any other path text, and `resolveURL`'s subsequent
`url.Parse` correctly splits the resolved string back into path+query), **brands**
(`create_brand`/`update_brand`, `POST`/`PUT /ref/brand`), and **payment terms**
(`create_payment_term`/`update_payment_term`, `POST`/`PUT /ref/paymentterm`). Every write is an
**external mutation requiring approval** (`metadata.json`'s `risk.write`): each directly creates or
overwrites a live Cin7 Core record referenced by other live records (a renamed category/brand/
payment-term is visible on every product/customer/supplier that already references it by name).

Legacy implemented none of these — it was read-only end to end (`Write` unconditionally returned
`connectors.ErrUnsupportedOperation`). See Known limits for the writes/deletes this pass
deliberately did NOT model, and `api_surface.json` for the full researched exclusion list.

## Known limits

- **`accountid` as a secrets-store alias is not modeled.** Legacy's `cin7AccountID` allows the
  account id to arrive via EITHER `cfg.Config["accountid"]` OR `cfg.Secrets["accountid"]`
  (secrets checked first) — some deployments store the (non-secret) account id alongside real
  secrets for convenience. This bundle declares `accountid` as a `spec.json` `config` property
  only; a caller that previously supplied it exclusively via the secrets store must move it to
  config. This is a config-surface narrowing (ACCEPTABLE, never changes accepted-input DATA,
  conventions.md §5) — the resolved account id value once supplied is identical either way.
- **`id`-field fallback chains are narrowed to the primary field only.** Legacy's `firstField`
  helper falls back through a priority list when the primary id-shaped field is absent:
  `products.id` = `firstField(item, "ID", "SKU")`, `sale_list.id` = `firstField(item, "SaleID",
  "ID")`, `purchase_list.id` = `firstField(item, "ID", "TaskID")`. The engine's `computed_fields`
  dialect has no "first of N paths" coalesce primitive — only a single bare `{{ record.<path> }}`
  reference (or a filter chain over one reference). This bundle references only the
  higher-priority field (`ID`/`SaleID`/`ID` respectively); the documented Cin7 Core wire shape
  always populates that field for every real product/sale/purchase record (the fallback exists
  defensively in legacy for malformed/partial API responses, never observed in practice).
  ACCEPTABLE per conventions.md §5's meta-rule: this never diverges for any record the real Cin7
  API actually returns; it is a genuine `ENGINE_GAP` only for the theoretical malformed-response
  case legacy defends against.
- `page_size` (the query size param) is a fixed `streams.json` pagination value (100, matching
  legacy's default); legacy's config-driven `page_size` override (1-1000) is not modeled —
  the engine's `page_number` paginator's `PageSize` is a static value set once in `streams.json`,
  not template-resolvable (same shape as the aha/appfigures wave2 precedent). `max_pages` (legacy's
  0/all/unlimited-or-positive-integer request-count cap) is likewise not modeled — `spec.json`
  intentionally omits both `page_size` and `max_pages` (a declared-but-unwireable key is worse than
  an absent one, per conventions.md F6).
- Legacy's `base_url` SSRF-guard scheme/host validation (https/http only, host required) is
  reproduced by the engine's own base-URL handling; no bundle-level behavior change.
- **Deletes are modeled only for `product_categories`.** `product`/`customer`/`supplier`/
  `ref/brand`/`ref/paymentterm`/`ref/location`/`ref/carrier`/`ref/account`/`ref/unit` all expose a
  documented `DELETE`, but each is excluded as `destructive_admin` in `api_surface.json` (breadth-
  vs-cost triage on the riskiest mutation per resource) rather than modeled as a write action —
  every one of them is irreversible and several (location, chart-of-accounts, unit of measure) can
  be referenced by a large number of existing live records.
- **Chart of Accounts and Tax Rule writes (`POST`/`PUT`) are excluded as `requires_elevated_scope`/
  `out_of_scope`.** Cin7's own docs note Chart-of-Accounts updates are disabled outright while a
  Xero/QuickBooks accounting integration is enabled, and both resources carry live
  tax-calculation/accounting-posting configuration shared by every future transaction — genuinely
  higher-stakes than the 6 resource groups this pass covers writes for.
- **The Sale/Purchase/Advanced-Purchase workflow resource groups (quote → order → fulfilment →
  invoice → credit note → payment) are entirely out of scope**, both as new streams and as writes.
  `sale_list`/`purchase_list` already give full summary-level read parity with legacy; the
  underlying `Sale`/`Purchase`/`Advanced Purchase` detail resources are deeply nested, multi-stage
  workflow state machines (each sub-endpoint mutates one lifecycle stage with real inventory-
  allocation and accounting-posting side effects) — a large, genuinely separate scope from this
  connector's catalog/customer/supplier/order-summary surface. See `api_surface.json` for the full
  enumerated exclusion list (Production, Disassembly, Finished Goods, Stock adjustments/takes/
  transfers, Money Task/Journal, and the bundled CRM module are excluded for the identical reason).
- **`product_availability`'s primary key assumes one row per (product, location, bin) combination**
  — the documented response does not state an explicit uniqueness constraint beyond "stock
  information," so `["id", "location", "bin"]` is this bundle's best-effort reconstruction from the
  documented field list and example responses (a product with no bin-level tracking returns `Bin:
  null` for every location row, which still uniquely identifies that location's stock line).
