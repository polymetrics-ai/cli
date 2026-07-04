# Overview

e-conomic is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the FULL practical
surface of the legacy REST API at `https://restapi.e-conomic.com` (the `X-AppSecretToken`/
`X-AgreementGrantToken`-authenticated product; a separate, newer `apis.e-conomic.com` OpenAPI
product line is out of scope, see `api_surface.json`). It reads customers, products, suppliers,
accounts, booked and draft invoices, draft and archived orders, and reference/lookup data
(customer/product/supplier groups, payment terms, VAT zones, currencies). It also now WRITES:
customer/product/supplier create-update-delete, draft-invoice create/update, and booking a draft
invoice into a final booked invoice. This bundle migrates `internal/connectors/e-conomic` (the
hand-written connector, which was read-only); the legacy package stays registered and unchanged
until wave6's registry flip — the write surface here is a genuinely NEW Pass-B capability, not a
migrated legacy behavior, since legacy itself never wrote.

## Auth setup

Provide two secrets: `app_secret_token` (the e-conomic app's secret) and `agreement_grant_token`
(the per-agreement grant). Both are sent as plain request headers —
`X-AppSecretToken: <app_secret_token>` and `X-AgreementGrantToken: <agreement_grant_token>` — via
`streams.json`'s `base.headers`, matching legacy's `requester` construction exactly
(`e_conomic.go`'s `appSecretHeader`/`agreementGrantHead` constants). Both secrets are required;
neither has a fallback. `base_url` defaults to `https://restapi.e-conomic.com` and may be
overridden for tests/proxies. The same two secrets authenticate every write action — e-conomic has
no separate write-scoped credential.

## Streams notes

All 12 streams (`customers`, `products`, `suppliers`, `accounts`, `invoices`, `invoices_drafts`,
`customer_groups`, `product_groups`, `supplier_groups`, `payment_terms`, `vat_zones`, `currencies`,
`orders_drafts`, `orders_archived`) share the same list shape: `GET` against the e-conomic
collection path, records at `collection`, and e-conomic's own `skippages`/`pagesize` pagination
convention with `pagination.nextPage` (an absolute URL) driving the next page (`pagination.type:
next_url`, `next_url_path: "pagination.nextPage"`) — matching legacy's `harvest` function exactly
for the 5 original streams, and applying the identical documented convention to every newly-added
stream (confirmed via `restdocs.e-conomic.com` and e-conomic's own TechTalk developer-blog posts on
`/orders` and invoice-drafts/booking). `invoices` reads from `/invoices/booked`; `invoices_drafts`
reads from `/invoices/drafts` (unbooked, still-mutable work-in-progress invoices).
`orders_drafts`/`orders_archived` read from `/orders/drafts` and `/orders/archived` respectively —
e-conomic's `/orders` API is itself split into a mutable `drafts` collection and an immutable
`archived` collection once an order is sent (mirroring the invoices drafts/booked split). Every
stream's initial request sends `skippages=0&pagesize=100` (legacy's `defaultPageSize`);
`page_size`/`max_pages` are not modeled as runtime config (see Known limits).

Every stream requires a `computed_fields` rename because e-conomic's wire shape uses camelCase
field names (`customerNumber`, `salesPrice`, `creditLimit`, ...) while this bundle's schemas use
snake_case, and schema-as-projection matches only by exact raw-key equality (conventions.md §2) —
a plain projection with no renames would silently drop every e-conomic field. Nested business-key
references (e.g. `{"vatZone":{"vatZoneNumber":1}}`) are flattened via dotted `record.<path>`
references (`{{ record.vatZone.vatZoneNumber }}`), matching legacy's `refNumber` helper — a bare
single reference like this receives the engine's typed extraction (conventions.md §3), preserving
the real integer type, exactly like legacy's `refNumber` returning the raw `any` value. The 6
reference-data streams (`customer_groups`/`product_groups`/`supplier_groups`/`payment_terms`/
`vat_zones`/`currencies`) are narrow lookup tables (id/name(/code) shape); `currencies`' primary
key is its ISO code (`code`, a string), not a numeric id.

## Write actions & risks

10 write actions, all newly added in this Pass-B expansion (legacy itself never wrote):

- `create_customer` / `update_customer` / `delete_customer` — full CRUD against `/customers`.
  `update_customer` is a `PUT` (e-conomic's full-resource-replace semantics: omitted optional
  fields may be cleared, not left untouched) scoped by `path_fields: ["customerNumber"]`.
  `delete_customer` is idempotent (`missing_ok_status: [404]`); e-conomic itself additionally
  rejects (409, surfaced as a normal write failure, not specially handled) deleting a customer with
  existing booked entries — this bundle does not attempt to special-case that response.
- `create_product` / `update_product` / `delete_product` — identical full-CRUD shape against
  `/products`, keyed by `productNumber` (a string, e-conomic's product numbers are not
  guaranteed-numeric).
- `create_supplier` / `update_supplier` / `delete_supplier` — identical full-CRUD shape against
  `/suppliers`, keyed by `supplierNumber`.
- `create_draft_invoice` / `update_draft_invoice` — author a draft (unbooked, still work-in-progress)
  invoice against `/invoices/drafts`; `update_draft_invoice` only succeeds while the invoice remains
  a draft (e-conomic rejects an update once booked).
- `book_invoice` — `POST /invoices/booked` with a `draftInvoice` reference (and optional
  `bookWithNumber`/`sendBy`, e.g. `sendBy: "ean"` for electronic NemHandel/EAN invoicing, per
  e-conomic's own TechTalk documentation), transitioning a draft into a legally-binding, thereafter
  immutable booked invoice. This is the highest-risk action in this bundle: e-conomic's own
  documentation states core invoice fields cannot be changed post-booking — a correction requires
  issuing a credit note against the booked invoice, not a further update/delete. Treat this action
  as effectively irreversible.

Orders (`orders_drafts`/`orders_archived`) remain READ-ONLY this wave — e-conomic's own
`/orders/drafts` POST/PUT/DELETE and `/orders/sent` transition are real, documented endpoints (see
`api_surface.json`) but are excluded as `out_of_scope`: legacy never modeled order writes at all,
and this Pass-B expansion's write scope was bounded to the invoice/customer/product/supplier
mutation surface; a future wave can add order-authoring writes following the identical
draft/sent-transition shape already proven by the invoice drafts/booked pair.

## Known limits

- **No incremental cursor on any stream.** Legacy exposes no incremental cursor field for any
  e-conomic stream (`economicStreams()` declares no `CursorFields` anywhere), and none of the
  newly-added Pass-B streams introduce one either — e-conomic's REST API supports `filter`/`sort`
  query parameters but no server-side "modified since" semantics this bundle wires up. Every stream
  is full-refresh only; no `incremental` block is declared anywhere, matching legacy's original 5
  streams and extending the same shape to the 7 new ones.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, capped at 1000) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`pageSize`/`maxPages` in `e_conomic.go`). The engine's `next_url` paginator has no
  config-driven page-size or max-pages knob, so this bundle sends legacy's own default
  (`pagesize=100`) as a static per-stream query literal and does not declare `page_size`/`max_pages`
  in `spec.json` at all (a declared-but-unwireable config key is worse than an absent one, per
  conventions.md F6 precedent). Pagination is bounded only by e-conomic's own empty-`nextPage` stop
  signal, matching e-conomic's real termination behavior.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance
  in the legacy Go connector) stamps deterministic placeholder records with fields shaped after —
  but not identical to — the live wire shape. This bundle's schemas and fixtures target the LIVE
  record shape only (`e_conomic.go`'s `harvest`/`mapRecord` functions), per the bitly-pilot
  precedent (`docs/migration/conventions.md`'s worked example): the engine's own
  fixture-replay conformance harness supersedes the need for an in-connector fixture-mode branch.
- **`next_url` pagination ships single-page conformance fixtures for every stream** (the sanctioned
  exception, conventions.md §4): e-conomic's `pagination.nextPage` is an absolute URL whose host is
  the live API (or, in a real sync, whatever `base_url` resolves to) — a static fixture file cannot
  embed the conformance replay server's own dynamically-assigned address ahead of time. All 12
  streams share the identical base-level `next_url` pagination, so every stream fixture here is
  single-page; `pagination_terminates` exercises `customers` (the first declared stream) against
  its one-page fixture, which trivially proves the read terminates and consumes exactly the one
  recorded page. A live two-page proof (an `httptest.Server` asserting the engine correctly follows
  a `nextPage` URL across two real pages) is out of scope for this wave (JSON+docs only, no
  `paritytest` packages) and is a documented follow-up: extend this bundle with a
  `paritytest/e-conomic` suite (mirroring bitly's/calendly's `next_url` parity tests) in a
  subsequent wave.
- **The newer `apis.e-conomic.com` OpenAPI product line is out of scope.** e-conomic separately
  publishes versioned OpenAPI products (`customersapi`, `journalsapi`, `subscriptionsapi`, ...) at
  `apis.e-conomic.com` with their own base URLs (e.g.
  `https://apis.e-conomic.com/customersapi/v3.1.0/`) while sharing the same two auth headers. This
  bundle's `spec.json`/`streams.json` target the legacy `restapi.e-conomic.com` surface only, the
  one legacy's own hand-written connector was built against; the newer product line is a distinct
  API surface with its own versioning lifecycle and is not modeled here (see `api_surface.json`'s
  final entry).
- **Write actions have no dedicated hook/compound logic.** Every write action here is a single
  plain HTTP request (`body_type: "json"`, default body construction) — none needed a `WriteHook`.
  `book_invoice`'s real-world usage often chains a preceding `create_draft_invoice` (or a prior
  `update_draft_invoice`) followed by this action with the resulting `draftInvoiceNumber`; this
  bundle does not compose that chain automatically (each action is one independent write call, per
  the engine's per-record write model) — the caller is responsible for sequencing draft-then-book
  across two separate write calls.
