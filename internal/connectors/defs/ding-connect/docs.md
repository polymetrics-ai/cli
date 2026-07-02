# Overview

Ding Connect is a wave2 migration of `internal/connectors/ding-connect` (the
hand-written legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip). It reads DingConnect's
read-only reference/catalog data — countries, currencies, regions,
providers, and products — through the DingConnect REST API. It is
read-only: DingConnect's source surface here exposes no reverse-ETL writes,
so `capabilities.write` is `false` and this bundle ships no `writes.json`.

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

All 5 streams share DingConnect's uniform list-endpoint envelope: `GET`
against `/api/V1/<Resource>`, records extracted from the response's
top-level `Items` array. Pagination is `offset_limit` with `offset_param:
Skip` and a static `page_size: 100` — DingConnect list endpoints accept no
server-side page-size query parameter at all (`limit_param` is intentionally
unset, so the engine never sends one), matching legacy's `harvest()`
(`ding_connect.go:138-170`) exactly: `Skip` starts at 0 and advances by 100
on every full (100-record) page; a short page (or an empty one) stops
pagination.

DingConnect's reference resources carry no natural id; the upstream API
assigns none, so legacy derives a synthetic `uuid` primary key by joining
select fields with `:` (`dingUUID`, `ding_connect.go:303-311`). Every stream
reproduces this exactly via `computed_fields`: `countries`/`currencies`/
`providers`/`products` key on a single field (`CountryIso`/`CurrencyIso`/
`ProviderCode`/`SkuCode` respectively — legacy's own single-key `dingUUID`
calls for these), and `regions` joins two fields
(`{{ record.CountryIso }}:{{ record.RegionCode }}`), matching legacy's
`dingUUID(item, "CountryIso", "RegionCode")` call for that stream.

## Write actions & risks

None. Legacy `ding-connect` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full DingConnect API surface (balance queries, top-up/transfer mutations)
  is out of scope for wave2; see `api_surface.json`'s `excluded` entries.
  Only the 5 legacy-parity reference/catalog read streams are implemented.
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
