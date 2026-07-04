# Overview

Pennylane started as a read-only declarative migration of `internal/connectors/pennylane` (legacy
Go connector) and has been expanded to broad documented API-surface coverage in Pass B. It reads
customers, customer invoices, suppliers, supplier invoices, products, categories, transactions, and
bank accounts, and writes company/individual customer, supplier, product, and category mutations
through Pennylane's External API v2 (`https://app.pennylane.com/api/external/v2`, 159 documented
endpoints as of this review). Legacy only ever implemented the original 5 read streams and exposed
no writes at all; `supplier_invoices`/`transactions`/`bank_accounts` and every write action are new
Pass-B capability, not a legacy port — legacy stays registered and unchanged until wave6's registry
flip.

## API surface (Pass B)

`api_surface.json` now enumerates all 159 documented endpoints (retrieved from
`https://pennylane.readme.io/llms.txt` and each endpoint's individual reference page, which embeds
the real OpenAPI operation). 8 streams + 9 write actions are `covered_by`; every other endpoint
carries a specific `excluded` category:

- `duplicate_of` — single-resource-by-id GETs and per-parent sub-resource lists (invoice lines,
  appendices, matched transactions, categories-of-X, etc.) that this bundle's list streams already
  return in full, or for which there is no dynamic-id enumeration source (no `fan_out` config-list
  or parent-listing request wired) to reach the sub-resource at all.
- `requires_elevated_scope` — real, tracked capability gaps: entire resource domains not yet
  modeled (ledger/journals — under a documented 2026 scope migration — billing_subscriptions,
  quotes, purchase_requests, payment mandates), and compound/lifecycle writes materially more
  complex than this bundle's existing single-object create/update actions (nested invoice-line
  creation, invoice finalize/mark-paid/e-invoice-status transitions, transaction-to-invoice
  reconciliation/matching, category-assignment sub-writes).
- `binary_payload` — every multipart file-upload endpoint (invoice/quote/purchase-order imports,
  file/appendix attachments); this dialect's `body_type` is `json`/`form`/`none` only.
- `non_data_endpoint` — async export-job endpoints, the trial balance report, and `/me` (token/
  company metadata, not a syncable collection).
- `out_of_scope` — pure notification/email-send side effects (send invoice/quote by email, mandate
  signature-request emails) with no reverse-ETL data-mutation analog.
- `deprecated` — the one endpoint (`ledger_attachments` upload) the docs themselves mark deprecated
  in favor of `file_attachments`.

## Auth setup

Provide a Pennylane API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. Legacy hard-errors when `api_key` is
unset (`pennylane connector requires secret api_key`), matching this bundle's `required:
["api_key"]`.

## Streams notes

All 8 streams (`customers`, `customer_invoices`, `suppliers`, `products`, `categories`,
`supplier_invoices`, `transactions`, `bank_accounts`) share the identical shape: `GET` against the
Pennylane v2 list endpoint, records at `items`, primary key `["id"]`. Pagination matches legacy's
cursor loop: `pagination.type: cursor` with `cursor_param: cursor` and `token_path: next_cursor`;
the bundle follows `next_cursor` until it is absent/empty. Fixtures still include Pennylane's real
`has_more` response field, but it is not used as a separate stop signal because legacy did not use
it. `limit` is sent on every request from `page_size` (default `50`, matching legacy's
`defaultPageSize` for the 5 legacy streams; the 3 new Pass-B streams use the same shared default).

The 5 legacy streams declare `x-cursor-field: updated_at`, matching legacy's published
`CursorFields`; the 3 new Pass-B streams do not declare cursor metadata.

All 8 streams declare `"projection": "passthrough"` (post-wave2 review §8 rule 1): none of them is
built from a field-by-field `mapRecord`-style function, so schema-mode projection would silently
drop any raw field the schema omits. Each schema documents a representative subset of Pennylane's
real wire fields as a documentation surface only; it does not gate what is emitted.

Legacy also forwards two optional, verbatim passthrough config values as query params whenever
set: `filter` and `sort` (`if filter := ...; filter != "" { base.Set("filter", filter) }` /
same for `sort`). This bundle wires both through the engine's opt-in optional-query dialect
(`query.filter`/`query.sort` object-form entries with `omit_when_absent: true`) — present only
when the corresponding config value is set, absent entirely otherwise, matching legacy's
conditional `url.Values.Set` calls exactly. Every real Pennylane list endpoint documents its own
`filter`/`sort` field vocabulary (e.g. `supplier_invoices` supports filtering by
`id`/`supplier_id`/`invoice_number`/`date`/`category_id`/`external_reference`/`payment_status`/
`flow_id`) — this bundle passes the caller's filter/sort expression through verbatim rather than
validating it against each endpoint's specific field list, matching legacy's own pass-through
behavior for the 5 original streams.

## Write actions & risks

Pass B adds `create_product` (Pennylane documents this as a plain, uncomplicated single-object
create — unlike the customer-invoice/quote create endpoints, which require nested line items and
are excluded as `requires_elevated_scope` in `api_surface.json`) alongside the 8 pre-existing write
actions (`create_company_customer`/`update_company_customer`/`create_individual_customer`/
`update_individual_customer`/`create_supplier`/`update_supplier`/`update_product`/`create_category`/
`update_category`). `metadata.json` now declares `capabilities.write: true` (previously incorrectly
left `false` even though `writes.json` already carried the original 8 actions — a stale leftover
from before this bundle's write surface was added, fixed as part of this pass, mirroring the same
fix applied to pabbly-subscriptions-billing in this same Pass B increment).

## Known limits

- `x-cursor-field: updated_at` is declared on `customers`/`customer_invoices`/`suppliers`/
  `products`/`categories` as catalog/candidate-cursor metadata only, mirroring legacy's own
  `CursorFields: []string{"updated_at"}` declaration for the 5 original streams. No stream applies
  any server-side or client-side incremental filter using this field (no `updated_at[gte]`-style
  query param, no client-side filtering by cursor value), so no `streams.json` `incremental` block
  is declared.
- `page_size` config validation (legacy's numeric range) is not reproduced at the bundle-config
  level; the engine treats `page_size` as an opaque string substituted directly into the `limit`
  query param. This never changes emitted record DATA for any legacy-valid input; it only narrows
  client-side input validation.
- Legacy also accepts a runtime `max_pages` cap, but the declarative engine only supports fixed
  bundle-authored `pagination.max_pages` integers. This bundle intentionally does not declare an
  ignored `max_pages` `spec.json` property.
- Every real Pennylane endpoint documents required OAuth2 scopes (e.g. `customers:all`,
  `supplier_invoices:readonly`) that a live API key must actually hold; this bundle's Bearer-token
  auth has no scope-awareness of its own — an API key missing a required scope surfaces as an
  ordinary 401/403 from the live API, not a bundle-side pre-check. This matches every other
  declarative bundle's auth model (the engine has no OAuth-scope-declaration primitive) and is not
  Pennylane-specific.
- The `requires_elevated_scope`-excluded resource domains (ledger/journals, billing_subscriptions,
  quotes, purchase_requests, payment mandates) and compound/lifecycle writes are real, tracked
  capability gaps — see `api_surface.json` for the full per-endpoint reasoning — not silently
  narrowed scope.
