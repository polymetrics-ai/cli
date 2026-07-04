# Overview

Visma e-conomic is a Danish/Nordic accounting SaaS. This bundle reads customers, suppliers,
products, booked and draft invoices, draft and sent orders, draft and sent quotes, departments,
payment terms, units, VAT types, VAT zones, accounts, customer groups, and product groups from the
e-conomic REST API (`{base_url}`), and writes customers, suppliers, products, units, and payment
terms. It migrates `internal/connectors/visma-economic` (the hand-written legacy connector, which
stays registered and unchanged until wave6's registry flip) and, per Pass B, expands well beyond
legacy's single read-only `customers` stream to the connector's full declared core-commercial
surface — see `api_surface.json` for the complete endpoint-by-endpoint accounting.

## Auth setup

Two secrets are required: `app_secret_token` and `agreement_grant_token`, e-conomic's own
two-token app authentication scheme. Both are sent as static request headers
(`X-AppSecretToken` / `X-AgreementGrantToken`) on every request via `streams.json`'s `base.headers`
— e-conomic does not use a Bearer/Basic/API-key-query scheme, so `base.auth` is declared as a single
unconditional `{"mode": "none"}` (the credentials flow entirely through the two headers, not
through the `auth` dispatch). Both secrets are required in `spec.json`; an absent header-templated
secret is always a hard validate/runtime error (per the engine's header-resolution rule), matching
legacy's own `Check`/`requester` validation that rejects an empty `app_secret_token` or
`agreement_grant_token`.

`base_url` defaults to `https://restapi.e-conomic.com`, matching legacy's `defaultBaseURL` constant,
materialized via `spec.json`'s `"default"` value.

## Streams notes

All 17 streams share the same envelope, pagination, and incremental shape:

- **Records envelope**: every list endpoint returns its collection under a top-level `collection`
  array (e-conomic's uniform list envelope) — `records.path: "collection"` on every stream.
- **Pagination**: `base.pagination` declares `page_number` with `page_param: skippages`,
  `size_param: pagesize`, `start_page: 0` (e-conomic's `skippages` is 0-indexed — the number of
  PAGES to skip, not a record offset), `page_size: 100`. This applies uniformly to every stream;
  none override it.
- **Incremental filtering**: e-conomic supports a uniform `lastUpdated$gte:<ISO8601>` query filter
  (its own documented `filter` parameter grammar) on every resource whose object carries a
  `lastUpdated` field. Streams whose real wire objects expose `lastUpdated`
  (`customers`/`suppliers`/`products`/`invoices_booked`/`invoices_drafts`/`orders_drafts`/
  `orders_sent`/`quotes_drafts`/`quotes_sent`) declare `incremental.cursor_field: lastUpdated` +
  `request_param: filter`, gated by `omit_when_absent` (§3's optional-query dialect) so a full-refresh
  read sends no `filter` param at all — matching e-conomic's own "omit for unfiltered" contract, not
  an empty-string filter. Static reference/master-data collections with no `lastUpdated` field on
  their own object (`departments`/`payment_terms`/`units`/`vat_types`/`vat_zones`/`accounts`/
  `customer_groups`/`product_groups`) declare no `incremental` block — there is no cursor to expose.
- **id derivation**: every stream stamps `id` via `computed_fields` from that resource's own
  `<resource>Number` field (e-conomic's own primary-key convention — `customerNumber`,
  `supplierNumber`, `productNumber`, `bookedInvoiceNumber`, `draftInvoiceNumber`, `orderNumber`,
  `quoteNumber`, `departmentNumber`, `paymentTermsNumber`, `unitNumber`, `vatTypeNumber`,
  `vatZoneNumber`, `accountNumber`, `customerGroupNumber`, `productGroupNumber`). `customers` keeps
  legacy's exact stringified-id behavior via the `last_path_segment` filter (forces string output,
  reproducing legacy's `fmt.Sprint(item["customerNumber"])`); every other stream uses a bare
  `{{ record.<field> }}` reference, which the engine's typed-extraction rule copies through as the
  raw JSON type (an integer for every numeric `<resource>Number`, a string for `products`'
  alphanumeric `productNumber`) — this is a genuine, honest type per resource, not a parity
  requirement, since none of these streams existed in legacy to have an established output type.
- `orders_drafts`/`orders_sent` and `quotes_drafts`/`quotes_sent` are modeled as four separate
  streams (not one `orders`/`quotes` stream with a status field) because e-conomic itself exposes
  them as two structurally separate top-level collections (`/orders/drafts` vs `/orders/sent`,
  `/quotes/drafts` vs `/quotes/sent`) with independent pagination/incremental state — matching the
  API's own resource boundaries, the same reasoning `invoices_booked`/`invoices_drafts` already
  follow.

## Write actions & risks

15 actions across 5 resources — every one a real, external mutation requiring approval:

- **Customers**: `create_customer` (POST `/customers`, requires `name`/`currency`/`paymentTerms`/
  `customerGroup`/`vatZone` per e-conomic's own required-field contract), `update_customer` (PUT
  `/customers/{id}`), `delete_customer` (DELETE, `missing_ok_status: [404]` — idempotent delete).
- **Suppliers**: `create_supplier`, `update_supplier`, `delete_supplier` — identical shape to
  customers, using supplier's own `group`/`paymentTerms`/`vatZone` required references.
- **Products**: `create_product` (requires `productNumber`/`name`/`salesPrice`/`productGroup`),
  `update_product`, `delete_product`.
- **Units**: `create_unit`, `update_unit`, `delete_unit` — e-conomic's simplest master-data
  resource, a bare `name` string.
- **Payment terms**: `create_payment_term`, `update_payment_term`, `delete_payment_term` — requires
  `name`/`paymentTermsType` (e-conomic's own terms-type enum, e.g. `netCash`, `currentMonth`).

**Not migrated** (see `api_surface.json` for the full per-endpoint accounting): invoice/order/quote
draft creation and update (`POST`/`PUT /invoices/drafts`, `/orders/drafts`, `/quotes/drafts`) require
composing a line-items sub-array (product/quantity/account references) with e-conomic's own
totals/VAT computation contract — a shape this dialect's flat-record write body cannot express
without risking a request that silently diverges from what a real e-conomic client sends;
`POST /invoices/booked` (booking a draft) is a two-step compound sequence (fetch a
booking-instructions template, then POST it) that needs a WriteHook, out of this connector's Pass B
scope. Department/customer-group/product-group create/update mutations are lower-priority
master-data writes left for a future wave.

## Known limits

- Invoice/order/quote line items are read-only structure within each record, not an independently
  writable sub-collection — see "Write actions & risks" above.
- Booking a draft invoice, and moving a draft order/quote to sent status, are not migrated (both
  are compound multi-request sequences beyond a single declarative write action) — see
  `api_surface.json`'s `requires_elevated_scope` entries for `POST /invoices/booked`,
  `POST /orders/sent`, `POST /quotes/sent`.
- Accounting-ledger-internal resources (accounting years, journal entries, vouchers, layouts,
  templates, employees, app-roles, projects, cost-types, payment-types, currencies) are
  intentionally out of scope — this connector's declared surface is customers/suppliers/products/
  invoicing/quoting plus the master-data collections those resources directly reference, not
  general-ledger bookkeeping.
- No parent-scoped sub-resource streams (customer contacts, customer delivery locations,
  per-customer/per-group invoice/customer sub-listings) — each would need a `fan_out` block keyed
  to every already-read customer/group id, which is a heavier read pattern than this pass's core
  list-endpoint scope; the equivalent unfiltered top-level collection (e.g. `invoices_booked`) is
  already covered.
- PDF rendering (`GET /invoices/booked/{n}/pdf`) and file-attachment upload
  (`POST /orders/{n}/attachment/file`) are binary-payload endpoints this dialect's JSON-only
  schema/write model cannot express.
