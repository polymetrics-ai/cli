# Overview

Adobe Commerce (Magento) is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the
full practically-syncable Magento REST surface. It reads Magento products, orders, customers,
categories, invoices, shipments, credit memos, customer groups, and store website/store-view
configuration through the Magento REST API (`GET https://<store>/rest/V1/...`), and writes product
updates, category create/update, and order cancellation. This bundle originally targeted capability
parity with `internal/connectors/adobe-commerce-magento` (the hand-written connector it migrates,
which is read-only); Pass B's full-surface research (`api_surface.json`) goes beyond that legacy
parity baseline per docs/migration/conventions.md's Pass B scope. The legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Magento Integration Access Token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and never logged, matching legacy's `connsdk.Bearer(secret)`
(`adobe_commerce_magento.go:277`). `base_url` is required and must be the fully composed Magento
REST base URL including the API version prefix (e.g. `https://magento.mystore.com/rest/V1`) — see
Known limits for why this diverges from legacy's `store_host`+`api_version` derivation.

## Streams notes

Ten streams total. `products`, `orders`, `customers`, `categories`, `invoices`, `shipments`,
`creditmemos`, and `customer_groups` share the identical Magento `searchCriteria` list shape:
`GET /<resource>` returns `{"items":[...],"total_count":N,"search_criteria":{...}}`, records live at
the `items` array. `customers`, `categories`, and `customer_groups` read from Magento's dedicated
search/list endpoints (`/customers/search`, `/categories/list`, `/customerGroups/search`) exactly
like legacy's `magentoStreamEndpoints` routing table extended to the new resources.
`store_websites`/`store_views` (`/store/websites`, `/store/storeViews`) are unpaginated,
top-level-array store-configuration endpoints (`pagination.type: none`, `records.path: "."`) — small,
bounded, admin-scoped lists with no `searchCriteria` support at all.

Pagination for the `searchCriteria`-shaped streams is `page_number`: `searchCriteria[currentPage]`
(1-based, `start_page: 1`) and `searchCriteria[pageSize]` (static `page_size: 100`, matching legacy's
`magentoDefaultPageSize`). The engine's `page_number` paginator stops on a short page
(`recordCount < page_size`); legacy additionally checks `total_count` and stops as soon as the
accumulated count reaches it. These two stop conditions are equivalent for every dataset except one
whose total record count is an exact multiple of the page size, where legacy stops immediately via
the `total_count` comparison and the engine would issue one additional request that returns an empty
page before stopping — no different records are ever emitted either way, so this is a
request-count-only, non-data-affecting divergence (documented here per the conventions.md §5
meta-rule, since it is a real behavioral difference even though not a data one).

`products`/`orders`/`customers`/`categories`/`invoices` are incremental on `updated_at`, matching
legacy's `incrementalLowerBound` (state cursor, falling back to `start_date`). `shipments` and
`creditmemos` are incremental on `created_at` instead — Magento's Sales-module shipment/credit-memo
list endpoints do not reliably index an `updated_at` field the way orders/invoices do (both are
effectively append-only once created), so `created_at` is the correct real-wire filterable
timestamp for these two streams; this is new (Pass B) coverage with no legacy precedent to diverge
from. Legacy expresses the `<field> > lower_bound` filter as Magento's three-part
`searchCriteria[filter_groups][0][filters][0][field/value/condition_type]` query convention; every
incremental stream reproduces it via three `stream.Query` entries, each gated on
`{{ incremental.lower_bound }}` with `omit_when_absent: true` so all three are present together on
an incremental read and absent together on a full sync (`field`/`condition_type` use the `const:`
filter to send their fixed literals — `updated_at`/`created_at` and `gt` — only when the lower bound
itself resolves; `value` sends the lower bound's own formatted value). No `incremental.request_param`
is declared since Magento's filter has no single param name to hold — the three `stream.Query`
entries carry the whole filter instead. `customer_groups`/`store_websites`/`store_views` declare no
`incremental` block: customer groups and store configuration are small, slow-changing admin lists
with no meaningful cursor field in Magento's own schema.

## Write actions & risks

Four write actions, all requiring approval (external mutation of a live Magento store):

- **`update_product`** (`PUT /products/{sku}`, `path_fields: ["sku"]`): partial product field update
  (name/price/status/visibility/weight). Magento's PUT semantics merge the supplied fields onto the
  existing product rather than replacing it wholesale, matching this dialect's default
  except-path-fields body construction.
- **`create_category`** (`POST /categories`): creates a new catalog category under a required
  `parent_id`.
- **`update_category`** (`PUT /categories/{id}`, `path_fields: ["id"]`): partial category field
  update (name/is_active/position).
- **`cancel_order`** (`POST /orders/{id}/cancel`, `path_fields: ["entity_id"]`, `body_type: "none"`):
  irreversibly cancels a live sales order; Magento's cancel endpoint takes no request body, only the
  order id in the path.

Product/order/category **creation via a full raw payload**, invoice/shipment/credit-memo creation,
and any DELETE endpoint are excluded — see `api_surface.json` for the full reasoned exclusion list
(compound multi-part payloads this dialect's flat single-record body construction cannot safely
express, and destructive/elevated-scope admin actions this connector does not allow-list).

## Known limits

- **`base_url` requires the fully composed REST base URL; the `store_host`+`api_version` derivation
  is not modeled.** Legacy derives its base URL from a `store_host` config value plus an optional
  `api_version` override, defaulting to `V1`, at request time (`adobe_commerce_magento.go:304-338`,
  `magentoBaseURL`). The engine's `spec.json` `"default"` materialization mechanism
  (conventions.md §3) only fills in a FIXED literal default, not one derived from another config
  value — there is no declarative way to express "concatenate `store_host` + `/rest/` +
  `api_version`" without inventing ad hoc Go (a Tier-2 escalation this bundle does not need for
  anything else). This bundle therefore requires `base_url` directly, already including the
  `/rest/V1` (or other version) suffix; an operator who previously configured `store_host` must now
  configure the fully composed URL instead. This is a documented config-surface narrowing per
  conventions.md §3's "derived default" guidance, not a silent behavior change to any request this
  bundle actually sends once configured.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-300,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`magentoPageSize`/`magentoMaxPages`). The engine's `page_number` paginator reads its
  page size only from `PaginationSpec.PageSize`, a bundle-declared static integer with no
  template/config wiring (conventions.md's rate_limit/dead-config precedent applies identically
  here — see stripe's resolved ledger item 3), and there is no config-driven request-count cap
  mechanism at all for the `page_number` paginator. `page_size`/`max_pages` are therefore not
  declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, F6); this
  bundle sends Magento's own effective default (`searchCriteria[pageSize]=100`) as a static
  pagination-block value, matching stripe's `limit=100` precedent.
- **`total_count`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 100, never a difference in which records are emitted.
- **Multi-Source Inventory, CMS pages/blocks, sales rules/coupons, and tax configuration are out of
  scope.** MSI (`/inventory/sources`, `/inventory/stocks`, `/inventory/source-items`) is an Adobe
  Commerce-edition-only module, not guaranteed present on a Magento Open Source install this
  connector otherwise targets uniformly; CMS content and merchandising/tax configuration are
  distinct product domains from the catalog/sales/store-config surface this bundle covers. See
  `api_surface.json` for the full per-endpoint reasoning.
- **Quote/cart (checkout-flow) endpoints are out of scope.** An active Magento cart/quote is a
  session-scoped, in-progress object with no stable server-assigned id suitable for a syncable read
  stream, and its mutation surface is an inherently multi-step, stateful checkout sequence (create
  cart, add items, set addresses, place order) rather than a single-record write. The `orders`
  stream already covers the settled result once a cart becomes an order.
- **Order/invoice/shipment/credit-memo creation are not modeled as writes.** Each of these real
  Magento endpoints requires a compound payload (order line items, per-item invoice/ship/refund
  quantity breakdowns) that this dialect's flat, single-record write-body construction cannot safely
  express without risking data corruption on the live store; only the simpler, already-well-defined
  `cancel_order` state-transition action and catalog product/category field updates are modeled.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`) stamps deterministic synthetic records with a `previous_cursor`
  field echoing `req.State["cursor"]` when set (`adobe_commerce_magento.go:253-255`). This is a
  credential-free conformance-harness affordance with no live-path equivalent; this bundle's
  schemas and fixtures target the live record shape only, and the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
