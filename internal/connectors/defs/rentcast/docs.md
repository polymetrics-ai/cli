# Overview

RentCast is a wave2 fan-out declarative-HTTP migration. It reads RentCast properties, sale
listings, rental listings, market data, and value/rental estimates through the RentCast REST API
(`GET https://api.rentcast.io/v1/...`). This bundle targets capability parity with
`internal/connectors/rentcast` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a RentCast API key via the `api_key` secret; it is sent as the `X-Api-Key` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-Api-Key", key, "")`, and
is never logged. `base_url` defaults to `https://api.rentcast.io/v1` (legacy's
`rentcastDefaultBaseURL`) and can be overridden for tests or proxies.

## Streams notes

All six streams (`properties`, `sale_listings`, `rental_listings`, `markets`, `value_estimates`,
`rental_estimates`) return a bare JSON array (`records.path: ""`), matching legacy's
`connsdk.RecordsAt(resp.Body, "")`. Pagination is `offset_limit` (`limit`/`offset`, static
`page_size: 100` matching legacy's `rentcastDefaultPageSize`) with the identical short-page stop
rule legacy's own `harvest` implements (`len(records) < pageSize` stops the loop) — this is an
exact parity match, not an approximation.

Legacy applies five optional config-driven filters (`address`, `city`, `state`, `zipCode`,
`propertyType`) uniformly to **every** stream's request (`rentcastFilters(cfg)`), regardless of
whether that stream's endpoint documents support for the filter. This bundle reproduces that exact
behavior via the opt-in optional-query dialect (`query.<param>.omit_when_absent: true`), declared
identically on **each of the six streams' own `query` block** — `engine.HTTPBase` has no `query`
field at all, so a `base.query` block is silently dropped by the JSON decoder (an unknown key, not
an error) and would never reach an outgoing request; `engine.StreamSpec.Query` is the only field
the engine actually reads (`docs/migration/conventions.md`'s optional-query dialect is a
per-stream primitive, not a base-level one). Each stream therefore repeats the same five-filter
`query` object so every filter is sent only when its corresponding config value
(`address`/`city`/`state`/`zip_code`/`property_type`) is set, and omitted entirely otherwise —
identical to legacy's `if value != "" { query.Set(...) }` loop, and identical across streams
because legacy applies the same filter set uniformly regardless of per-endpoint support.

Each stream's schema is a field-for-field projection of legacy's own `mapRecord` functions.
`properties`/`sale_listings`/`rental_listings` rename RentCast's camelCase wire fields
(`formattedAddress` -> `address`, `propertyType` -> `property_type`, `zipCode` -> `zip_code`,
`lastSeenDate` -> `last_seen_date`) via `computed_fields`, matching legacy's `first(item,
"formattedAddress", "address")`-style helpers for the PRIMARY (first-preference) source field.
`markets` renames `zipCode` -> `zip_code`. All three listing/property streams publish
`last_seen_date` as `x-cursor-field`, matching legacy's own `CursorFields` declarations, but
RentCast's endpoints expose no server-side incremental filter parameter and legacy's own `harvest`
never applies one — every read is a full paginated sweep, matching legacy's true read behavior
exactly (no `request_param`/`start_config_key`/`client_filtered` declared). `markets` has no cursor
field, matching legacy.

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's multi-key fallback chains for derived fields are narrowed to their primary source
  field only.** Legacy's `first(item, keys...)` helper tries each key in order and uses the first
  non-nil value, e.g. `id` on `markets` falls back to `first(item, "id", "zipCode")`, and
  `value_estimates`/`rental_estimates`'s `price`/`rent` fall back to `first(item, "price", "value")`
  / `first(item, "rent", "value")`. The `computed_fields` dialect resolves a single bare template
  reference (or a literal) per output field with no OR-fallback primitive across multiple record
  paths, so this bundle wires only the PRIMARY (first-listed, most commonly populated) source key
  for each field: `markets.id` reads raw `id` only (no `zipCode` fallback when `id` is absent);
  `value_estimates.price`/`rental_estimates.rent` read `price`/`rent` only (no `value` fallback).
  This is a documented scope narrowing, not a silent divergence: for any response where the primary
  key is present (the common case for RentCast's own documented wire shape), behavior is identical
  to legacy; a response relying on legacy's secondary fallback key would silently drop that field
  here instead. `properties`/`sale_listings`/`rental_listings`'s `address`/`property_type` also use
  only the primary key (`formattedAddress`/`propertyType`) since RentCast's real API consistently
  uses those field names — the secondary alias (`address`/`propertyType`-alternate) exists in legacy
  purely as defensive handling for a shape RentCast's real API does not appear to emit.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `offset_limit` paginator has no config-driven page-size or
  request-count-cap knob (mirrors this wave's aha/referralhero precedent); `page_size`/`max_pages`
  are therefore not declared in `spec.json`, and this bundle sends RentCast's own default
  (`limit=100`) as a static pagination-block value.
- Full RentCast API surface (short-term rental listings, liens, bulk data exports, tax history) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
