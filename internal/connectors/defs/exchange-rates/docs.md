# Overview

Exchange Rates API (exchangeratesapi.io) is a read-only foreign-exchange rate data API. This
bundle migrates only the `latest` stream from the legacy `internal/connectors/exchange-rates`
package to a Tier-1 defs bundle. The legacy `exchange_rates` (daily historical) and `symbols`
streams are NOT ported here — see Known limits.

## Auth setup

Provide an exchangeratesapi.io access key via the `access_key` secret; it is sent as the
`access_key` query parameter on every request (never logged). `base_url` defaults to
`https://api.exchangeratesapi.io/v1` and only needs overriding for tests or proxies. An optional
`base` config value (the source base currency) is appended to the `latest` request when set, and
omitted entirely when absent (matching legacy's `baseQuery` helper).

## Streams notes

- `latest` (`GET /latest`) — a single-object response (`records.path: "."`,
  `single_object: true`), primary key `date`. The nested `rates` object survives schema
  projection unmodified (no `computed_fields` needed — `records.rates` is already the correct
  output field name and shape; the engine's schema-mode projection copies a matched-key nested
  object through by reference, exactly matching legacy's `normalizeRates`/`rateRecord` mapping for
  every response that actually includes a `rates` object).

## Write actions & risks

None. Exchange Rates API is read-only in both legacy and this bundle (`capabilities.write:
false`, no `writes.json`).

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
- **`symbols` is not ported (blocked, ENGINE_GAP).** The `/symbols` response body is a
  `{"symbols": {"USD": "United States Dollar", "EUR": "Euro", ...}}` map; legacy explodes this
  into one `{code, name}` record per map key (sorted for determinism). The declarative
  `records.path` extraction (`connsdk.RecordsAt`) supports only an array-of-objects shape or a
  single-object passthrough (`single_object: true`) — there is no "treat each key of this object
  as a separate record" mode. Expressing this correctly needs either a new engine records-mode or
  a `StreamHook`; both are out of scope for this wave.
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here (matches legacy's actual behavior).
- The `latest` stream's `rates` field is passed through as-is (whatever shape the API returns); a
  malformed/missing `rates` object is emitted as absent rather than legacy's defensive
  `normalizeRates` fallback to `{}` — this only differs for a response shape the real API does
  not produce for the `latest` endpoint, so it is an ACCEPTABLE deviation (never diverges for any
  legacy-accepted real response).
