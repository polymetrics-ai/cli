# Overview

Chargify (Maxio Advanced Billing) is a wave2 fan-out Tier-1 declarative migration, expanded in Pass
B to the full practical API surface. It reads Chargify customers, subscriptions, products, product
families, coupons, transactions, invoices, payment profiles, events, and statements through the
Chargify REST API, and writes customers, subscriptions (create/update/cancel), product families,
products, and coupons. This bundle originally targeted capability parity with
`internal/connectors/chargify` (the hand-written connector it migrates, read-only); the legacy
package stays registered and unchanged until wave6's registry flip. Pass B's write actions are new
capability beyond legacy parity — see "Write actions & risks" below.

## Auth setup

Chargify authenticates with HTTP Basic. By default, provide a Chargify API key via the `api_key`
secret; it is sent as the Basic username with the literal password `"x"` (Chargify's documented
API-key auth convention — `chargify.go`'s `chargifyBasicPassword`), matching
`auth: [..., {"mode":"basic","username":"{{ secrets.api_key }}","password":"x","when":"{{
secrets.api_key }}"}]`. Legacy also accepts an explicit `username` (config) + `password` (secret)
pair that takes precedence over the API key when BOTH are set
(`chargifyCredentials`: `if username != "" && password != "" { return username, password, nil }`
checked before the api_key fallback) — this bundle reproduces that exact precedence with the
username/password candidate declared FIRST in `base.auth` (conventions.md §3's dual-auth-ordering
rule), gated `"when": "{{ config.username }}"`. The API host is `base_url`, which is **required** in
this bundle (see "Known limits" below for why).

**Narrow auth-gating deviation (documented, not a blocker)**: the engine's `when` grammar rejects
`&&`/`||` (`EvalWhen`, conventions.md §3), so a single candidate cannot gate on "both `username` AND
`password` are set" the way legacy's `username != "" && password != ""` check does. This bundle
gates the username/password candidate on `config.username` truthiness alone. For the two realistic
configurations (both set; neither set) behavior is identical to legacy. The one degenerate corner
case this cannot reproduce exactly is `username` set with `password` unset: legacy falls through to
the `api_key` candidate in that case, whereas this bundle would select Basic auth with an empty
password (and then hard-error, since `secrets.password` is unresolved and `password` is a plain
`{{ }}` reference, not `when`-gated — any templated AuthSpec field other than `when` still hard-errors
on an absent secret, per §3's header/field resolution rules applied identically to auth fields).
An operator who sets `username` is, by construction, opting into the explicit-override pair and is
expected to set `password` alongside it; this is judged ACCEPTABLE scope-narrowing (an
unambiguously-misconfigured corner case fails loudly rather than silently degrading) rather than an
`ENGINE_GAP` blocker.

## Streams notes

All 5 streams (`customers`, `subscriptions`, `products`, `coupons`, `transactions`) share the same
shape: `GET` against the Chargify list endpoint (`/customers.json` etc.), records at the response's
top-level JSON array (`records.path: "."`), each element wrapped in a single-key resource envelope
(e.g. `{"customer": {...}}`) exactly like legacy's `unwrap()` helper. Pagination follows Chargify's
`page`/`per_page` convention (`pagination.type: page_number`, `page_param: page`,
`size_param: per_page`, `start_page: 1`, `page_size: 100` — matches legacy's
`chargifyDefaultPageSize`); a page shorter than the declared `page_size` ends the read, matching
legacy's `harvest()` loop (`chargify.go:152-183`) exactly. Primary key is `["id"]`
across every stream; the incremental cursor field is `["updated_at"]` for `customers`/
`subscriptions`/`products`/`coupons` and `["created_at"]` for `transactions`, matching legacy's
`chargifyStreams()` catalog. **No `incremental.request_param`/`client_filtered` is declared for any
stream** — legacy itself never filters server-side or client-side by the cursor (its own comment:
"Chargify list endpoints do not accept an updated_at lower bound across all streams"; `InitialState`
only tracks the cursor "for resumability", never uses it to gate a request), so every read is a full
refresh regardless of any persisted cursor, exactly matching legacy. `x-cursor-field` is still
declared on each schema (satisfying the engine's cursor-field-exists validation and enabling
`*_deduped`/`incremental_append` sync modes at the catalog level) without changing any request
behavior.

**Envelope unwrap via per-field `computed_fields`** (conventions.md §2 schema-as-projection,
chargebee's golden pattern): Chargify wraps every list item in a single-key resource envelope, so
plain schema projection (which looks up each schema property directly on the raw extracted record)
would see only that one wrapper key and produce an empty record. Every schema property is therefore
populated by a `computed_fields` entry reaching into the envelope (e.g. `"id": "{{
record.customer.id }}"`, `"created_at": "{{ record.customer.created_at }}"`), matching legacy's
`chargifyCustomerRecord` (and its 4 sibling `mapRecord` functions in `streams.go`) field-for-field —
including TYPE: every computed_fields entry here is a single bare `{{ record.<envelope>.<field> }}`
reference with no filter stage, so the engine's typed computed_fields extraction copies the raw JSON
value straight through (numeric/boolean fields preserve their native type instead of being
stringified). Schemas declare the real wire type (`integer`/`boolean`/`string`) per field, matching
`chargifyStreamEndpoints`'s field catalog exactly.

**Pass B new streams** (same envelope-unwrap pattern as above): `product_families`
(`/product_families.json`, envelope `product_family`), `invoices` (`/invoices.json`, response shape
`{"invoices": [...], "meta": {...}}` — NOT the single-key-per-element envelope the other streams
use, since Chargify's newer relational Invoicing API returns whole invoice objects directly, so
`invoices`' `computed_fields` reference `record.<field>` directly rather than
`record.invoices.<field>`; `records.path: "invoices"` selects the array), `payment_profiles`
(`/payment_profiles.json`, envelope `payment_profile`, no incremental cursor declared — Chargify's
payment-profile list has no documented updated_at-ordered pagination guarantee and no legacy
precedent to match), `events` (`/events.json`, envelope `event`, cursor `created_at` — events are
an immutable append-only log), and `statements` (`/statements.json`, envelope `statement`, no
incremental — Chargify's statement list has no documented cursor semantics).

## Write actions & risks

Pass B adds write capability (new beyond legacy, which was read-only): `create_customer`/
`update_customer`, `create_subscription`/`update_subscription`/`cancel_subscription`,
`create_product_family`, `create_product`/`update_product`, `create_coupon`/`update_coupon`.

**Every write action's `record_schema` nests its mutable fields under a single wrapper property
matching Chargify's own request-envelope convention** (`{"customer": {...}}`,
`{"subscription": {...}}`, `{"product_family": {...}}`, `{"product": {...}}`, `{"coupon": {...}}` —
Chargify's real API requires exactly this envelope on every write body, confirmed against
`docs.chargify.com`'s signup/customer examples). The write dialect's default JSON body construction
(`buildJSONBody`, `engine/write.go`) copies every top-level record field verbatim into the request
body except `path_fields` — it has no dedicated "wrap the body under this key" mechanism, so the
wrapper key is modeled as an ordinary nested-object `record_schema` property instead (the same
pattern airtable's `create_record`'s `fields` object uses): a record shaped
`{"id": "1", "customer": {"first_name": "Jane"}}` produces exactly Chargify's expected
`PUT /customers/1.json` body `{"customer": {"first_name": "Jane"}}`, with `id` excluded from the
body by `path_fields`. Subscription creation only models the minimal `product_handle` +
`customer_id` shape (Chargify's full signup payload also accepts nested `payment_profile`/
`components`/`metafields` sub-objects); a caller needing those can still pass them as additional
keys inside the `subscription` object, since the schema does not set `additionalProperties: false`.

`cancel_subscription` (`POST /subscriptions/{id}/cancel.json`) sends an optional
`{"subscription": {"cancellation_message": "..."}}` body — Chargify accepts cancellation without a
message too, so `subscription` is not required.

All write actions carry `"risk": "external mutation; approval required"` (subscription lifecycle
actions specifically call out billing side effects) — `metadata.json`'s `capabilities.write` is now
`true` and `risk.write` documents the aggregate exposure.

## Known limits

- **`components` (`GET /product_families/{id}/components.json`) is an `ENGINE_GAP`, not
  implemented**: this endpoint is scoped per product-family, requiring a fan-out over every
  `product_families` id. Chargify wraps each `product_families.json` list element in a singular
  envelope (`{"product_family": {"id": ...}}`), but `fan_out.ids_from.request.id_field`
  (`engine/read.go`'s `fanOutIDsFromRequest`) does a single flat `map[string]any` key lookup
  (`rec[spec.IDField]`) with no dotted-path traversal — `id_field` can select a top-level `"id"` key
  but not a nested `"product_family.id"` path, so against Chargify's real envelope shape zero ids
  would ever resolve and the stream would silently emit nothing. This was not shipped as a
  silently-broken fan_out; see `api_surface.json`'s `ENGINE_GAP`-tagged entry. Closing this requires
  either a dotted-path-capable `id_field` or a `records_path`/pre-projection step that can reach into
  the wrapper before extraction.
- **Product/coupon/component price points, per-subscription component allocation, and per-product
  detail sub-resources are out of scope** — see `api_surface.json`'s per-endpoint `out_of_scope`
  reasons (mostly: nested per-parent-id sub-resource collections with no top-level list endpoint,
  deferred alongside the `components` gap above rather than half-modeled against a single hardcoded
  parent id).
- Payment-profile create/update is excluded (`requires_elevated_scope`): Chargify steers raw
  card-data capture through Chargify.js/hosted tokenization rather than a plain server-side JSON
  POST, which this connector's declarative write dialect has no mechanism to represent safely.
  Payment profiles remain read-only in this bundle.
- Destructive/irreversible admin actions (customer/coupon/payment-profile delete, subscription
  purge, invoice status override) are deliberately excluded — see `api_surface.json`'s
  `destructive_admin` entries.
- **`domain`/`subdomain` config keys dropped; `base_url` is now required.** Legacy derives the API
  host from a `domain` (or `subdomain` + `.chargify.com`) config value when `base_url` is unset
  (`chargifyBaseURL`). The engine's spec-default materialization only fills in a LITERAL per-key
  default — it cannot express "derive `base_url` from `domain`/`subdomain`", a cross-key template
  (the same class chargebee's `site`/sentry's `hostname` hit). Per `docs/migration/conventions.md`'s
  guidance for this exact shape, this bundle drops `domain`/`subdomain` entirely and requires
  `base_url` instead: an operator migrating a legacy `domain`/`subdomain`-only config must now supply
  the fully-formed `https://<domain>.chargify.com` URL as `base_url`. This is a documented
  config-surface narrowing (every legacy-accepted `domain`/`subdomain` value has an
  operator-reachable `base_url` equivalent; no request/data change once configured), not a
  data-shape regression.
- **`page_size`/`max_pages` config keys dropped.** `streams.json`'s `pagination.page_size` is a
  static JSON int (`PaginationSpec.PageSize`), not a runtime-templated value — there is no engine
  mechanism to let a `spec.json` config property override it per-read. Legacy defaults `page_size`
  to 100 (`chargifyDefaultPageSize`, configurable up to 200 via a `page_size` config value); this
  bundle declares the identical `page_size: 100` static default (restored — an earlier draft of
  this bundle had leaked a fixture-sized `page_size: 2` into the live `streams.json` pagination
  block, which would have shipped a 50x-smaller production page size than legacy's default; the
  required 2-page fixture (conventions.md §4) instead ships 100 full customer records on page 1 and
  1 record on page 2 to prove pagination continuation at the real page size). Legacy's
  `page_size`/`max_pages` config properties are consequently genuinely dead config in this dialect
  and are not declared in `spec.json` (F6, conventions.md — a declared-but-unwireable key is worse
  than an absent one).
- `metadata.json` declares no `rate_limit` block: legacy Chargify enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chargify.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule.
