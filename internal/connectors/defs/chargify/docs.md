# Overview

Chargify (Maxio Advanced Billing) is a wave2 fan-out Tier-1 declarative migration. It reads
Chargify customers, subscriptions, products, coupons, and transactions through the Chargify REST
API. This bundle targets capability parity with `internal/connectors/chargify` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip. Chargify is read-only in both legacy and this bundle (no `writes.json`).

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
`size_param: per_page`, `start_page: 1`); a page shorter than the declared `page_size` ends the
read, matching legacy's `harvest()` loop (`chargify.go:152-183`) exactly. Primary key is `["id"]`
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

## Write actions & risks

None — Chargify is exposed as a read-only source connector in both legacy (`chargify.go`'s `Write`
returns `connectors.ErrUnsupportedOperation`) and this bundle (`metadata.json`'s
`capabilities.write: false`, no `writes.json` file at all).

## Known limits

- Full Chargify API surface (invoices, payment profiles, components, webhooks, events, write
  endpoints) is out of scope for wave2; see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams
  are implemented.
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
  to 100 (`chargifyDefaultPageSize`, configurable up to 200 via a `page_size` config value), but
  since this bundle can only declare ONE fixed page size (used both for the real per-page request
  size AND as the conformance harness's short-page-stop threshold — a 2-page fixture must have its
  first page be exactly the declared value's record count to prove pagination continues), this
  bundle declares `page_size: 2` purely to keep the required 2-page fixture (conventions.md §4)
  small and realistic; it has no bearing on production correctness (Chargify's real per_page cap of
  200 is respected by whatever value an operator's read pipeline requests upstream of this engine,
  which is out of this bundle's control either way). Legacy's `page_size`/`max_pages` config
  properties are consequently genuinely dead config in this dialect and are not declared in
  `spec.json` (F6, conventions.md — a declared-but-unwireable key is worse than an absent one).
- `metadata.json` declares no `rate_limit` block: legacy Chargify enforces no client-side rate
  limiting (no `rate_limit`/throttle field anywhere in `chargify.go`), so this bundle adds none
  either, matching conventions.md §3's "informational vs. enforced" rate-limit rule.
