# Overview

Chargebee started as a wave1-pilot Tier-1 declarative migration (PLAN.md P-6, SPEC.md §5.4) and was
expanded to full API-surface coverage in Pass B (api_surface.json `reviewed_at: 2026-07-03`). It
reads 32 Chargebee resources (customers, subscriptions, invoices, plans, items, item prices, item
families, coupons, coupon codes/sets, credit notes, transactions, orders, quotes, payment sources,
events, hosted pages, virtual bank accounts, unbilled charges, ramps, gifts, alerts, comments,
promotional credits, features, entitlements, differential prices, price variants, products, webhook
endpoints, ledger operations, and ledger account balances) and writes 36 actions (the core
create/update/delete/void/cancel triad for its primary business objects) through the Chargebee v2
REST API (Product Catalog 2.0). The original 5 streams (`customers`, `subscriptions`, `invoices`,
`plans`, `items`) remain engine-vs-legacy parity-tested against `internal/connectors/chargebee` (the
hand-written connector this bundle originally migrated); the legacy package stays registered and
unchanged until wave6's registry flip, and is frozen at its original 5-stream, read-only surface —
it never gained the 27 additional streams or any write capability. See
`docs/migration/conventions.md` §8 for the Pass B review rules this expansion followed.

## Auth setup

Provide a Chargebee site API key via the `site_api_key` secret; it is used as the HTTP Basic
username with an empty password (`auth: [{"mode":"basic","username":"{{ secrets.site_api_key }}",
"password":""}]`), matching legacy's `connsdk.Basic(secret, "")` exactly (chargebee.go:262-264,278).
The API host is `base_url`, which is **required** in this bundle (`spec.json`'s
`required: ["site_api_key", "base_url"]`) — e.g. `https://{site}.chargebee.com/api/v2`. Legacy
instead DERIVES the host from a `site` config value (`chargebeeBaseURL`,
`"https://" + site + ".chargebee.com/api/v2"` when `base_url` is unset); this bundle does not
reproduce that derivation (see "Known limits" below for why and the config-surface change this
implies for an operator migrating a legacy-shaped config).

## Streams notes

Every stream shares the same base shape: `GET` against the Chargebee list endpoint, records at the
top-level `list` array, each element wrapped in a single-key resource envelope (e.g.
`{"customer": {...}}`, `{"item_price": {...}}`). Pagination follows Chargebee's `offset`/
`next_offset` convention uniformly across all 32 streams (`base.pagination.type: cursor` with
`cursor_param: offset` and `token_path: next_offset`): the next page's `offset` query value is read
verbatim from the previous response body's `next_offset` field, and pagination stops when
`next_offset` is absent. Every request sends `limit=100` via each stream's static
`query: {"limit": "100"}`.

**The original 5 legacy-parity streams** (`customers`, `subscriptions`, `invoices`, `plans`,
`items`) match legacy's `harvest()` loop (chargebee.go:148-196) exactly, including
`sort_by[asc]=updated_at` sent alongside `updated_at[after]` on every incremental request
(`chargebee.go:151-155`), expressed via the `incremental.lower_bound` query-var dialect (S3 engine
mini-wave item 1) — see "Known limits" below and `paritytest/chargebee/parity_test.go`'s
`TestParityChargebee_SortByAscSentOnIncrementalFromState`/
`TestParityChargebee_SortByAscSentOnIncrementalFromStartDate`/
`TestParityChargebee_SortByAscOmittedOnFullSync`.

**The 27 Pass B streams** were added directly from Chargebee's official OpenAPI spec
(`chargebee_api_v2_pc_v2_spec.json`, https://github.com/chargebee/openapi) rather than a legacy
connector (Chargebee's legacy Go package only ever implemented the original 5), so there is no
parity claim for them — only a documented-API-contract claim. Incremental coverage per stream
follows the truth table in conventions.md §8 rule 2 (`request_param` iff the API's own list endpoint
accepts an `updated_at`/`occurred_at` filter parameter, confirmed against the OpenAPI spec's
per-endpoint `parameters` list for every stream added):

- **Incremental via `updated_at[after]`** (server-side filter confirmed in the OpenAPI spec):
  `item_prices`, `item_families`, `credit_notes`, `transactions`, `orders`, `quotes`,
  `payment_sources`, `hosted_pages`, `virtual_bank_accounts`, `ramps`, `price_variants`, `products`.
  Of these, `coupons`/`credit_notes`/`quotes` support `updated_at[after]` filtering but their list
  endpoints do NOT accept `sort_by[asc]=updated_at` (only `created_at`/`date` sorting is offered —
  confirmed via the spec's per-endpoint `sort_by.properties.asc.enum`); the `sort_by[asc]` query
  key is therefore omitted entirely for those 3 streams' `query` blocks (declaring it would be a
  request Chargebee's own API rejects). Incremental filtering still returns exactly the records at
  or after the lower bound in both cases — only the ordering GUARANTEE within a page differs (a
  `sort_by`-less incremental page is not guaranteed strictly ascending by `updated_at`, but every
  record `>=` the lower bound is still returned across the full paginated read).
- **Incremental via `occurred_at[after]`** (Chargebee's events log has no `updated_at`, only an
  immutable `occurred_at`): `events`.
- **No incremental block** (the OpenAPI spec's list endpoint accepts no time-based filter parameter
  at all — confirmed absent from that endpoint's `parameters`): `coupon_codes`, `coupon_sets`,
  `unbilled_charges`, `gifts`, `alerts`, `comments`, `promotional_credits`, `features`,
  `entitlements`, `differential_prices`, `webhook_endpoints`, `ledger_operations`,
  `ledger_account_balances`. These are full-refresh-only streams; `x-cursor-field` is likewise
  absent from their schemas (conventions.md §8 rule 2: "neither → no incremental block").

**Non-standard primary keys**: `coupon_codes`' primary key is `["code"]` (Chargebee's own natural
key for that resource — coupon codes have no separate `id` field); `ledger_account_balances` has no
`id` field at all in the API response, so its primary key is the composite
`["subscription_id", "unit_id", "unit_type"]` the API itself uses to identify a balance row.

**Envelope unwrap via per-field `computed_fields`** (conventions.md §2 schema-as-projection):
Chargebee wraps every list item in a single-key resource envelope, so plain schema projection
(which looks up each schema property directly on the raw extracted record) sees only that one
wrapper key and produces an empty record. Every schema property is therefore populated by a
`computed_fields` entry reaching into the envelope (e.g. `"id": "{{ record.customer.id }}"`,
`"created_at": "{{ record.customer.created_at }}"`), matching legacy's `chargebeeCustomerRecord`
(and its 4 sibling `mapRecord` functions in streams.go) field-for-field — including TYPE, not just
value: every computed_fields entry here is a single bare `{{ record.<envelope>.<field> }}`
reference with no filter stage, so the engine's typed computed_fields extraction (gap-loop cycle-1
item 1) copies the raw JSON value straight through (numeric/boolean fields preserve their native
type instead of being stringified). Schemas declare the real wire type
(`integer`/`boolean`/`string`) per field, matching `chargebeeStreams()`'s field catalog exactly; see
`paritytest/chargebee/parity_test.go`'s
`TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes`.

## Write actions & risks

Pass B added write capability (`metadata.json`'s `capabilities.write` is now `true`); legacy
(`chargebee.go:258-260`'s `Write` still returns `connectors.ErrUnsupportedOperation` and is
unaffected by this bundle's expansion — see
`paritytest/chargebee/parity_test.go`'s `TestParityChargebee_LegacyWriteStillUnsupported`, which
pins that legacy behavior stays frozen, and `TestParityChargebee_CreateCustomerWriteSupported`,
which exercises the new capability end-to-end. 36 actions cover the core create/update/delete/
void/cancel triad for Chargebee's primary business objects, all `body_type: form` (matches every
mutation endpoint's documented `application/x-www-form-urlencoded` content type):

- **Customers**: `create_customer`, `update_customer`, `delete_customer`.
- **Items catalog**: `create_item`/`update_item`/`delete_item`, `create_item_price`/
  `update_item_price`/`delete_item_price`, `create_item_family`/`update_item_family`/
  `delete_item_family`.
- **Subscriptions**: `create_subscription` (POST `/customers/{id}/subscription_for_items`),
  `update_subscription` (POST `/subscriptions/{id}/update_for_items`), `cancel_subscription`
  (POST `/subscriptions/{id}/cancel_for_items` — irreversible; risk-flagged).
- **Billing documents**: `create_credit_note`/`void_credit_note`, `create_coupon`/`update_coupon`/
  `delete_coupon` (coupon creation/update route through Chargebee's Product-Catalog-2.0
  `create_for_items`/`update_for_items` endpoints — there is no plain `POST /coupons`),
  `create_order`/`update_order`/`cancel_order`, `void_invoice`, `collect_payment_for_invoice`
  (attempts to charge a payment method — risk-flagged).
- **Payments**: `create_card_payment_source` (carries raw card data via nested `card[...]`
  form fields — Chargebee's own form-encoding convention for the `card` object parameter;
  `write.go`'s `buildForm` sends record keys verbatim as form field names, so the record schema
  declares the bracketed key names directly, e.g. `"card[number]"`), `delete_payment_source`,
  `create_virtual_bank_account`/`delete_virtual_bank_account`.
- **Other**: `create_webhook_endpoint`/`update_webhook_endpoint`/`delete_webhook_endpoint`,
  `create_comment`/`delete_comment`, `add_promotional_credit`/`deduct_promotional_credit` (direct
  financial-credit effect — risk-flagged).

Every `delete`-kind action declares `delete.missing_ok_status: [404]` (an already-deleted record is
treated as successfully written, not failed), matching conventions.md §3's delete semantics.
Deliberately NOT covered as writes (see api_surface.json for the full, itemized exclusion list):
hard-deletes of invoices/credit notes/subscriptions (void/cancel are the safer, already-covered
reversible alternatives), quote/estimate/hosted-page checkout workflows (multi-step, no persisted
side effect until converted), invoice/credit-note payment-application and dunning-control actions,
and narrow catalog/packaging sub-resource management (differential prices, item entitlements,
price-variant attributes) beyond the core CRUD triads — breadth-first Pass B scope prioritizes real
business-object CRUD over exotic operational/admin actions.

## Known limits

- **Pass B full-surface expansion (this revision)**: `api_surface.json` was rewritten from the
  wave1-pilot's minimal-honest 13-endpoint manifest to a full enumeration of Chargebee's
  official Product-Catalog-2.0 OpenAPI spec (428 endpoints total, including the legacy `/plans`
  PC1.0 endpoint carried from the pre-Pass-B bundle). Every endpoint is `covered_by` a stream/write
  action XOR `excluded` with one of the closed-vocabulary categories
  (`destructive_admin`/`requires_elevated_scope`/`binary_payload`/`deprecated`/`non_data_endpoint`/
  `duplicate_of`/`out_of_scope`) and a specific, non-boilerplate reason — see `api_surface.json`'s
  `scope` field for the full breakdown. Notably excluded, with reasons: PDF/e-invoice generation and
  async bulk-export jobs (`binary_payload`); site/currency/custom-field configuration
  (`non_data_endpoint`); omnichannel/app-store billing, Product-Catalog-2.0 migration tooling, and
  multi-business-entity transfers, all gated behind add-ons most sites don't have enabled
  (`requires_elevated_scope`); the deprecated `/cards` endpoint (superseded by `payment_sources`)
  and short-lived payment tokens (`deprecated`/`duplicate_of`); irreversible hard-deletes and
  sandbox-only time-machine clock control, each with an already-covered safer alternative
  (`destructive_admin`); and compound checkout/quote/estimate workflows plus narrow catalog/
  packaging sub-resource management beyond the core CRUD triads (`out_of_scope`).
- **Chargebee's Product Catalog 1.0 vs 2.0**: the `plans` stream (legacy parity) reads the PC1.0
  `/plans` endpoint, which is NOT part of Chargebee's current public OpenAPI spec — it remains live
  for backward compatibility on sites that have not migrated to Product Catalog 2.0, but is
  undocumented in the current API reference. The 27 Pass B streams instead cover Chargebee's current
  PC2.0 surface (`items`/`item_prices`/`item_families` replace the PC1.0 `plans`/`addons` model for
  any site that has migrated). A site still on PC1.0 will return data for `plans` but likely empty
  or errored results for `item_prices`/`item_families`/`differential_prices`/`price_variants`
  (PC2.0-only concepts); this is an accurate reflection of Chargebee's own dual-catalog-version
  reality, not a bundle defect.
- Full Chargebee API surface (coupons, credit notes, addons, hosted pages, events, webhooks) is out
  of scope for wave1-pilot; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams are implemented.
  **SUPERSEDED by the Pass B full-surface expansion above** — kept for historical trace continuity;
  coupons/credit_notes/hosted_pages/events/webhook_endpoints are now implemented streams.
- **RESOLVED — computed_fields envelope unwrap now preserves native numeric/boolean types.**
  Previously (pre gap-loop-cycle-1), every schema field derived via a `computed_fields` envelope
  unwrap was stringified by `engine.Interpolate` regardless of the raw JSON value's real type,
  which forced every numeric/boolean schema property to a widened `["string", "null"]` type. The
  engine's typed computed_fields extraction (gap-loop cycle-1 item 1: a bare
  `{{ record.<path> }}` template with no filter stage copies the raw typed value instead of
  stringifying) now applies to every computed_fields entry in this bundle, so schemas declare the
  real wire type (`integer` for Unix-seconds timestamps and plain integers, `boolean` for booleans)
  matching legacy's `chargebeeStreams()` field catalog and `mapRecord` functions exactly, TYPE
  included. Asserted by `paritytest/chargebee/parity_test.go`'s
  `TestParityChargebee_ComputedFieldsPreserveNativeNumericAndBooleanTypes`.
  - **Why not a `RecordHook` instead** (SPEC §5.4's suggested fallback for cases computed_fields
    cannot reproduce exactly): `internal/connectors/conformance/dynamic.go`'s dynamic checks
    (`checkReadFixtureNonempty`, `checkPaginationTerminates`, `checkRecordsMatchSchema`,
    `checkCursorAdvances`) all call `engine.Read`/`engine.Check` with a literal `nil` Hooks
    parameter — a `RecordHook` would never fire during conformance, so `checkRecordsMatchSchema`
    would validate the schema against the still-envelope-wrapped raw record (one top-level key)
    instead of a flattened one, failing hard for every stream regardless of hook correctness.
    `computed_fields` is therefore the only mechanism whose output conformance actually exercises;
    with typed extraction it now ALSO preserves the real wire type, closing the gap this note
    originally documented. See `.planning/phases/wave1-pilot/traces/p6-chargebee-ledger.md` and
    `.planning/phases/wave1-pilot/traces/gaploop-s1-ledger.md`/`s2-chargebee-sentry-ledger.md` for
    the full design-decision trace.
- ~~**OPEN — `sort_by[asc]=updated_at` is not sent on incremental requests.**~~ **RESOLVED (S3 engine
  mini-wave item 1).** Legacy sets `sort_by[asc]=updated_at` alongside `updated_at[after]` on every
  incremental request whenever the computed lower bound is non-empty (`chargebee.go:151-155`), never
  on a full-refresh read. The engine now exposes the RESOLVED, post-`formatParam` incremental lower
  bound to `stream.Query` template resolution as `{{ incremental.lower_bound }}` (populated in
  `buildInitialQuery` BEFORE the query-template resolution loop runs, so it reflects EITHER the
  persisted `state.cursor` OR the `start_config_key` fallback — exactly the same value/precedence
  `updated_at[after]` itself uses). Composed with the existing `omit_when_absent` dialect and the new
  `const:<value>` filter (send a FIXED literal iff a reference resolves, without depending on the
  reference's own value), each stream's `query` now declares:
  ```json
  "sort_by[asc]": { "template": "{{ incremental.lower_bound | const:updated_at }}", "omit_when_absent": true }
  ```
  — present with the constant value `updated_at` iff the incremental lower bound resolves (state
  cursor or `start_date`), absent on a full-refresh read, exactly matching legacy's
  `if updatedAfter != ""` gate. See `paritytest/chargebee/parity_test.go`'s
  `TestParityChargebee_SortByAscSentOnIncrementalFromState`/
  `TestParityChargebee_SortByAscSentOnIncrementalFromStartDate`/
  `TestParityChargebee_SortByAscOmittedOnFullSync` and
  `.planning/phases/wave2-fanout-http-sm/traces/s3-engine-ledger.md` for the full design trace; the
  original STOP analysis remains at
  `.planning/phases/wave1-pilot/traces/s2-chargebee-sentry-ledger.md`'s chargebee item 2 section for
  historical reference.
- **`site` config key dropped; `base_url` is now required.** Legacy derives the API host from a
  `site` config value (`https://{site}.chargebee.com/api/v2`) when `base_url` is unset
  (`chargebeeBaseURL`). The engine's spec-default materialization (gap-loop cycle-1 item 6, C3)
  only fills in a LITERAL per-key default — it cannot express "derive `base_url` from `site`", a
  cross-key template. Per `docs/migration/conventions.md`'s guidance for this exact shape (sentry's
  `hostname` hit the identical class), this bundle drops `site` entirely and requires `base_url`
  instead: an operator migrating a legacy `site`-only config must now supply the fully-formed
  `https://{site}.chargebee.com/api/v2` URL as `base_url`. This is a documented config-surface
  narrowing (every legacy-accepted `site` value has an operator-reachable `base_url` equivalent; no
  request/data change once configured), not a data-shape regression.
- `metadata.json` declares no `rate_limit` block: legacy chargebee enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chargebee.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule (an absent
  block, not merely an unenforced one, since Chargebee's public rate limit was never documented in
  the legacy package to carry forward informationally).
