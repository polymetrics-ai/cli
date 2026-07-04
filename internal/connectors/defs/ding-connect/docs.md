# Overview

Ding Connect is a wave2 migration of `internal/connectors/ding-connect` (the
hand-written legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip), since Pass-B-expanded
to DingConnect's full practical V1 API surface. It reads DingConnect's
reference/catalog data — countries, currencies, regions, providers,
products, product descriptions, promotions, provider status, error code
descriptions — plus the distributor's live account balance, and sends
real-money mobile top-up transfers via `send_transfer`. `capabilities.write`
is `true` as of this Pass B expansion.

## Auth setup

Provide `api_key` as a secret. DingConnect authenticates with a bare
`api_key` request header (no `Bearer`/other prefix) — declared as `{"mode":
"api_key_header", "header": "api_key", "value": "{{ secrets.api_key }}"}`,
matching legacy's `connsdk.APIKeyHeader("api_key", secret, "")` construction
exactly (empty prefix).

`base_url` defaults to `https://api.dingconnect.com` (materialized via
`spec.json`'s `"default"`, matching legacy's `dingDefaultBaseURL`). An
optional `x_correlation_id` config value is sent as the `X-Correlation-Id`
header when set (declared but not `required`, so it is omitted entirely
when absent per conventions.md §3's conditional-header rule) — matches
legacy's `if corr := ...; corr != "" { headers["X-Correlation-Id"] = corr }`.

## Streams notes

The 5 legacy-parity streams share DingConnect's uniform list-endpoint
envelope: `GET` against `/api/V1/<Resource>`, records extracted from the
response's top-level `Items` array. Pagination is `offset_limit` with
`offset_param: Skip` and a static `page_size: 100` — DingConnect list
endpoints accept no server-side page-size query parameter at all
(`limit_param` is intentionally unset, so the engine never sends one),
matching legacy's `harvest()` (`ding_connect.go:138-170`) exactly: `Skip`
starts at 0 and advances by 100 on every full (100-record) page; a short
page (or an empty one) stops pagination.

DingConnect's reference resources carry no natural id; the upstream API
assigns none, so legacy derives a synthetic `uuid` primary key by joining
select fields with `:` (`dingUUID`, `ding_connect.go:303-311`). Every stream
reproduces this exactly via `computed_fields`: `countries`/`currencies`/
`providers`/`products` key on a single field (`CountryIso`/`CurrencyIso`/
`ProviderCode`/`SkuCode` respectively — legacy's own single-key `dingUUID`
calls for these), and `regions` joins two fields
(`{{ record.CountryIso }}:{{ record.RegionCode }}`), matching legacy's
`dingUUID(item, "CountryIso", "RegionCode")` call for that stream.

Pass B adds 5 more streams, all sharing the same `Items`-envelope shape but
declaring `pagination: {"type": "none"}` (stream-level override replaces the
base spec wholesale per conventions.md §3): DingConnect's own docs describe
these as small, non-paginated result sets (localized-string catalogs,
provider status, error taxonomy), unlike the potentially-large
countries/currencies/regions/providers/products catalogs.

- `product_descriptions` (`GET /api/V1/GetProductDescriptions`) — localized
  product display text, keyed by a synthetic
  `{{ record.LocalizationKey }}:{{ record.LanguageCode }}` `uuid` (this
  endpoint's own natural compound key, per DingConnect's docs: "always
  request the en language as well as the intended target language, as not
  all items will have translations in all languages").
- `promotions` (`GET /api/V1/GetPromotions`) — active provider promotions,
  keyed by a synthetic `{{ record.ProviderCode }}:{{ record.LocalizationKey
  }}:{{ record.StartUtc }}` `uuid` (no single natural id field is
  documented; this triple uniquely identifies one promotion window).
- `provider_status` (`GET /api/V1/GetProviderStatus`) — per-provider
  operational status (`IsProcessingTransfers`), keyed by `ProviderCode`.
- `error_code_descriptions` (`GET /api/V1/GetErrorCodeDescriptions`) — the
  static error-code taxonomy DingConnect's own responses reference, keyed by
  `Code`.
- `balance` (`GET /api/V1/GetBalance`) — the distributor account's live
  balance, a single JSON object (not an `Items` array), so this stream
  declares `records: {"path": ".", "single_object": true}` (the discord
  `guilds` stream's pattern) and stamps a static-literal `"balance"` `uuid`
  (there is only ever one balance record per sync).

## Write actions & risks

Legacy `ding-connect` was read-only; Pass B adds one write action:

- `send_transfer` (`POST /api/V1/SendTransfer`) — sends a mobile
  top-up/airtime transfer to a live account. Required fields: `SkuCode`
  (from the `products` stream), `SendValue`, `AccountNumber`,
  `DistributorRef` (the caller's own idempotency key); optional
  `SendCurrencyIso`, `ValidateOnly` (set `true` to validate without
  executing — DingConnect's own dry-run affordance). **Risk**: deducts the
  distributor's real DingConnect balance and initiates an actual top-up on
  the target account unless `ValidateOnly` is set; `metadata.json` marks
  `capabilities.write: true` and this action requires approval.

## Known limits

- DingConnect's full V1 surface (16 methods total) is now covered except 4
  endpoints; see `api_surface.json`'s `excluded` entries for exact reasons.
  `GetAccountLookup` is a per-account point lookup (requires a caller-chosen
  `account_number`), not a catalog listing with optional filters, so it does
  not fit the stream dialect's config/static-query-only shape.
- **`EstimatePrices`/`ListTransferRecords` are deliberately NOT
  implemented (`ENGINE_GAP`)**: both are read-shaped (no side effect,
  return computed/historical data) but require a POST body to carry their
  filter criteria; the engine's declarative read path (`read.go`'s
  `readDeclarative`) never sends a request body on any stream request
  (`Requester.Do(ctx, method, path, query, nil)` — always `nil`), so a
  body-carrying "read" cannot be expressed as a stream at all, regardless of
  HTTP method.
- **`CancelTransfers` is deliberately NOT implemented (`ENGINE_GAP`)**: it
  requires a top-level JSON **array** request body
  (`CancellationRequest[]`, each item a nested `{transfer_id: {transfer_ref,
  distributor_ref}, batch_item_ref}` object), but the write dialect's
  `executeWriteRecord` always issues one JSON **object** body per record
  (`buildJSONBody`/`buildBodyFieldsPayload` both return `map[string]any`,
  never array-wrapped) — there is no way to construct an array-shaped
  mutation body with the write dialect as it exists today.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`ding_connect.go:175-202`)
  emit synthetic records without any network call when `config.mode ==
  "fixture"` — this is a legacy-only testing convenience, not part of the
  live record shape; this bundle's own `fixtures/` directory is the wave2
  substitute used by `conformance`'s dynamic (fixture-replay) checks.
- **Documented deviation (`uuid` join, ACCEPTABLE)**: legacy's `dingUUID`
  helper skips any empty/missing component before joining with `:`, so a
  `regions` record missing `CountryIso` or `RegionCode` would legacy-side
  produce a shorter joined string (e.g. just `RegionCode` alone) rather than
  a literal `:`-prefixed/suffixed value. This bundle's `computed_fields`
  template (`"{{ record.CountryIso }}:{{ record.RegionCode }}"`) has no
  absent-field-skipping equivalent — a genuinely missing `CountryIso` would
  render as an empty string before the colon rather than being elided. Every
  real DingConnect region record carries both fields (they are the resource's
  own compound key), so this never diverges for any input the live API
  actually returns; recorded here per conventions.md §5's meta-rule since the
  dialect cannot express conditional field-skipping inside a computed field.
- **`page_size`/`max_pages` are not part of this bundle's config surface
  (documented scope narrowing, matching searxng's F6 precedent).** Legacy
  accepts both as config (clamped page size, page-count cap), but
  `PaginationSpec.PageSize`/`MaxPages` are static integer fields with no
  template support (conventions.md §3's pagination table) — there is no way
  to wire a runtime config value into either, so declaring them would be
  dead config a caller could set with no effect (F6, conventions.md).
  `page_size` is fixed at 100 in `streams.json`'s `base.pagination` block
  (matching legacy's own default), and DingConnect list endpoints accept no
  server-side page-size parameter at all either way.
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent
  DingConnect's real wire shape (`Items` envelope, PascalCase field names)
  exactly as the API returns them.
