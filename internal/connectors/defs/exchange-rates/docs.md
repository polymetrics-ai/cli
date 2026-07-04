# Overview

Exchange Rates API (exchangeratesapi.io) is a read-only foreign-exchange rate data API. This
bundle migrates the `latest` stream from the legacy `internal/connectors/exchange-rates` package
to a Tier-1 defs bundle, and adds 3 new streams in the Pass B full-surface expansion against the
real exchangeratesapi.io v1 docs (https://exchangeratesapi.io/documentation/): `convert`,
`timeseries`, and `fluctuation` — the API's complete documented surface beyond `latest` and the
two still-excluded streams below. The legacy `exchange_rates` (daily historical) and `symbols`
streams are NOT ported here — see Known limits (unchanged engine gaps, re-confirmed this pass).
exchangeratesapi.io has no write/mutation endpoints of any kind (a pure public FX-rate data API),
so `capabilities.write` stays `false` and this bundle ships no `writes.json`.

## Auth setup

Provide an exchangeratesapi.io access key via the `access_key` secret; it is sent as the
`access_key` query parameter on every request (never logged). `base_url` defaults to
`https://api.exchangeratesapi.io/v1` and only needs overriding for tests or proxies. An optional
`base` config value (the source base currency) is appended to the `latest`/`timeseries`/
`fluctuation` requests when set, and omitted entirely when absent (matching legacy's `baseQuery`
helper) — `convert` uses its own required `convert_from`/`convert_to` currency-code config instead
of `base`, since a currency conversion is inherently a from/to pair, not a base-currency rate
table.

## Streams notes

- `latest` (`GET /latest`) — a single-object response (`records.path: "."`,
  `single_object: true`), primary key `date`. The nested `rates` object survives schema
  projection unmodified (no `computed_fields` needed — `records.rates` is already the correct
  output field name and shape; the engine's schema-mode projection copies a matched-key nested
  object through by reference, exactly matching legacy's `normalizeRates`/`rateRecord` mapping for
  every response that actually includes a `rates` object).
- `convert` (`GET /convert`, Pass B addition) — a single-object response, primary key `date`
  (always present per the API's docs, including for a live/non-historical conversion). Requires
  `convert_from`/`convert_to`/`convert_amount` config (the API's required `from`/`to`/`amount`
  query params); `convert_date` is optional and omitted entirely when unset (`omit_when_absent`),
  giving a live-rate conversion rather than a historical one. The `query`/`info`/`result` fields
  survive schema projection unmodified as nested objects/scalars, matching the API's documented
  response shape exactly.
- `timeseries` (`GET /timeseries`, Pass B addition) — a single-object response, primary key
  `["start_date", "end_date"]` (the response is one record per requested window, not one record
  per date — the nested `rates` object itself contains one sub-object per date in the window,
  preserved as-is rather than fanned out into per-date records, matching `latest`'s "the nested
  object survives projection unmodified" precedent; fanning `rates` out into per-date records
  would need the same `keyed_object`-shaped mechanism `symbols` needs and hits an analogous
  narrower gap — see Known limits). Requires `timeseries_start_date`/`timeseries_end_date` config
  (the API's required `start_date`/`end_date` query params); `base` is optional or omitted.
- `fluctuation` (`GET /fluctuation`, Pass B addition) — a single-object response, primary key
  `["start_date", "end_date"]`, identical shape rationale to `timeseries` (the nested `rates`
  object holds one sub-object per currency code with `start_rate`/`end_rate`/`change`/
  `change_pct`, preserved as-is). Requires `fluctuation_start_date`/`fluctuation_end_date` config
  (the API's required `start_date`/`end_date` query params); `base` is optional or omitted.

## Write actions & risks

None. Exchange Rates API is read-only in both legacy and this bundle (`capabilities.write:
false`, no `writes.json`) — exchangeratesapi.io's full documented v1 surface (`symbols`, `latest`,
historical `/{date}`, `convert`, `timeseries`, `fluctuation`) has no POST/PUT/PATCH/DELETE
endpoint at all.

## Known limits

- **`exchange_rates` (daily historical) is not ported (blocked, ENGINE_GAP).** Legacy's
  `readExchangeRates` (`internal/connectors/exchange-rates/exchange_rates.go`) iterates one ISO
  date at a time from the incremental lower bound through `end_date` (or today), issuing up to
  366 separate `GET /<YYYY-MM-DD>` requests within a SINGLE `Read` call, skipping weekends when
  `ignore_weekends` is true. The declarative dialect's 6 pagination types (`none`, `link_header`,
  `page_number`, `offset_limit`, `cursor`, `next_url`) all advance the SAME endpoint path across
  pages; none can drive a per-call-varying PATH segment computed from a loop counter (a date
  sequence), and there is no "loop N times, changing the request path each time" primitive at all.
  This is a named `ENGINE_GAP`: reproducing it correctly (including the weekend-skip and
  366-day cap) requires a `StreamHook`, forbidden in this JSON-only wave.
- **`symbols` is still not ported (blocked, ENGINE_GAP) — the S4 `keyed_object` primitive does not
  cover this exact shape.** The `/symbols` response body is a `{"symbols": {"USD": "United States
  Dollar", "EUR": "Euro", ...}}` map; legacy explodes this into one `{code, name}` record per map
  key (sorted for determinism). The engine's `records.keyed_object`/`key_field` flatten (S4 engine
  mini-wave item 3, docs/migration/conventions.md §3 — added specifically to close appfigures'
  `{"111": {...}, "222": {...}}` map-of-OBJECTS shape) was evaluated and does NOT close this gap:
  `read.go`'s `recordsAtKeyed` requires every map VALUE to itself decode as a JSON object
  (`obj[k].(map[string]any)`) and silently SKIPS any key whose value is not an object — confirmed
  empirically (a `{"USD":"United States Dollar"}`-shaped body run through `recordsAtKeyed` emits
  zero records, since every value is a bare JSON string, not `{...}`). Exchange Rates' `/symbols`
  is a map of code → scalar display-name STRING, not a map of code → object — a narrower shape
  `keyed_object` does not support. Expressing this correctly still needs either a scalar-valued
  keyed-object mode (an engine extension distinct from the one that shipped) or a `StreamHook`;
  both remain out of scope for this wave.
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here (matches legacy's actual behavior).
- The `latest` stream's `rates` field is passed through as-is (whatever shape the API returns); a
  malformed/missing `rates` object is emitted as absent rather than legacy's defensive
  `normalizeRates` fallback to `{}` — this only differs for a response shape the real API does
  not produce for the `latest` endpoint, so it is an ACCEPTABLE deviation (never diverges for any
  legacy-accepted real response).
- The documented `latest` response does not include `historical`; legacy's field-built mapper would
  emit `"historical": null` for that absent key, while the engine omits absent fields. Keep fixtures
  faithful to the documented wire shape rather than adding `historical: false`.
- **`timeseries`/`fluctuation` do not fan their nested per-date/per-currency `rates` object out
  into individual records (documented scope narrowing, not an ENGINE_GAP blocker).** Both streams
  emit ONE record per requested `[start_date, end_date]` window, with the entire `rates` object
  (keyed by date for `timeseries`, by currency code for `fluctuation`) nested as-is, matching how
  `latest` already treats its own `rates` object. A caller wanting one row per date (timeseries) or
  per currency (fluctuation) would need the SAME `records.keyed_object` mechanism that cannot
  express `symbols`' scalar-valued map (`fluctuation`'s per-currency values ARE objects — `{
  start_rate, end_rate, change, change_pct }` — so `keyed_object` would actually apply there
  structurally; `timeseries`' per-date values are also objects, an inner code->number rate map).
  This bundle deliberately keeps both as a single nested-object record per window rather than
  reaching for `keyed_object` at the top level (which selects the OUTER stream-defining object,
  `rates` itself, not a nested field within an already-single-object response) — doing so would
  require restructuring `records.path` to point at `rates` and abandoning the `start_date`/
  `end_date`/`base`/`success` sibling fields entirely, a real information loss, not a pure
  cardinality change like `symbols`. Kept as ONE record per sync for both streams; a per-date/
  per-currency fan-out is a possible future enhancement, not a blocker.
