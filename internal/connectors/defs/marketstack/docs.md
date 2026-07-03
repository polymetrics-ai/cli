# Overview

Marketstack (`api.marketstack.com/v1`) is a read-only financial market data API. This bundle
migrates `internal/connectors/marketstack` (the hand-written legacy connector) to a declarative
Tier-1 bundle: 5 read-only streams (`exchanges`, `tickers`, `eod`, `splits`, `dividends`), no writes
— matching legacy's own read-only design ("no reverse-ETL surface makes sense" for market data).
Status: **partial** — see Known limits for one typed `ENGINE_GAP` affecting repeat incremental
syncs on the 3 date-cursored streams.

## Auth setup

Provide a Marketstack API access key via the `api_key` secret; it is sent as the `access_key` query
parameter on every request (`auth: [{"mode": "api_key_query", "param": "access_key", "value": "{{
secrets.api_key }}"}]`), matching legacy's `connsdk.APIKeyQuery("access_key", secret)` exactly. It
is never logged.

## Streams notes

- `exchanges` (primary key `["mic"]`) and `tickers` (primary key `["symbol"]`) are flat catalog
  streams with no incremental cursor, matching legacy's `marketstackStreams` entries with no
  `CursorFields`. Both flatten a nested object onto top-level schema fields via
  `computed_fields` dotted-path extraction: `exchanges` derives `currency_code`/`currency_name`/
  `currency_symbol` from the raw `currency{}` object and `timezone`/`timezone_abbr` from the raw
  `timezone{}` object; `tickers` derives `stock_exchange_mic`/`stock_exchange_name` from the raw
  `stock_exchange{}` object — matching legacy's `exchangeRecord`/`tickerRecord` nested-field
  flattening exactly (`docs/migration/conventions.md`'s dotted `record.<path>` reference).
- `eod`/`splits`/`dividends` (composite primary key `["symbol", "date"]`, incremental cursor field
  `date`) accept an optional `symbols` query filter (`config.symbols`, comma-separated tickers,
  `omit_when_absent: true` — sent only when configured, matching legacy's
  `if symbols := ...; symbols != "" { base.Set("symbols", symbols) }`) and a `date_from` lower bound
  computed from the incremental cursor or `start_date` (`incremental.request_param: date_from`,
  `param_format: date` truncates an RFC3339 value to `YYYY-MM-DD`, matching legacy's `dateOnly`
  helper). `exchanges`/`tickers` accept neither filter, matching legacy's `acceptsSymbols: false`
  for those two endpoints exactly.
- All 5 streams paginate with `pagination.type: offset_limit` (`limit_param: limit`, `offset_param:
  offset`), matching legacy's `harvest` loop, which advances `offset` by the page size and stops on
  a short page (`len(records) < pageSize`). **`page_size` is not exposed as config** —
  `PaginationSpec.PageSize` is a plain JSON int resolved once at bundle load, with no config-driven
  override in this engine version (the same static-pagination-field limitation documented in the
  auth0/searxng goldens, `docs/migration/conventions.md`). `streams.json`'s `pagination.page_size` is
  set to **100**, matching legacy's `marketstackDefaultPageSize` constant exactly — this static field
  is sent as `limit` on every request, fixture-replayed or live, so it must reflect legacy's real
  default rather than a small fixture-only value (a prior revision of this bundle shipped
  `page_size: 2`, which would have sent `limit=2` on every live request too; corrected per the wave2
  sweep's C3 finding). The required 2-page `pagination_terminates` proof (`exchanges` stream)
  therefore ships a 100-record page 1 (a full page, matching `page_size`) followed by a 1-record short
  page 2 — mirroring the auth0/aviationstack goldens' same page-size-realism-over-fixture-brevity
  tradeoff, not a shortcut.

Legacy enforces no client-side rate limiting, so this bundle declares no `streams.json`
`base.rate_limit` either, matching that (lack of) behavior exactly.

## Write actions & risks

None. Marketstack is a read-only source in both legacy and this bundle (`capabilities.write:
false`) — legacy's own `Write` method is an unconditional `ErrUnsupportedOperation` stub, and
market data has no sensible reverse-ETL surface.

## Known limits

- **`ENGINE_GAP` (repeat incremental sync on `eod`/`splits`/`dividends`)**: Marketstack's real wire
  format for the `date` cursor field uses a no-colon numeric UTC offset (`"2026-01-02T00:00:00+0000"`
  — confirmed by legacy's own fixture-mode synthetic record, `marketstack.go`'s `readFixture`, which
  reproduces this exact shape). A FRESH sync (lower bound sourced from the RFC3339 `start_date`
  config value, which is author-supplied and always carries a `Z`/colon-offset suffix) works
  correctly: `param_format: date` truncates it to `YYYY-MM-DD` via `time.Parse(time.RFC3339, ...)`
  with no issue. A REPEAT sync whose persisted state cursor is Marketstack's own emitted `date`
  value hits `engine/read.go`'s `parseLowerBoundTime`, which accepts only an all-digits Unix-seconds
  string or strict `time.RFC3339` (colon-delimited offset, e.g. `+00:00` or `Z`) — Go's
  `time.RFC3339` layout constant cannot parse a `+0000` (no colon) offset at all, so this parse
  hard-errors on a genuine repeat-sync cursor value derived from Marketstack's own response. Legacy
  never hit this because its own `dateOnly` helper is a pure string slice
  (`value[:strings.IndexAny(value, "T ")]`), never a `time.Parse` call, so it round-trips
  Marketstack's own timestamp shape unconditionally. Closing this requires either widening
  `parseLowerBoundTime` to also accept a non-colon numeric-offset RFC3339 variant (e.g. try
  `"2006-01-02T15:04:05-0700"` as a fallback layout) or a `param_format` variant that truncates via
  string slicing instead of full time parsing — either is an engine change, not a bundle-level
  workaround, since faking success here would mean silently sending a wrong/stale `date_from` on
  every second-and-later incremental sync. First-sync reads (the common bootstrap case) are
  unaffected. `conformance`'s `cursor_advances` dynamic check re-reads a stream seeded with the max
  cursor value it observes IN the fixture records themselves (Marketstack's real `+0000` wire
  shape) and requires the engine to format that value per `param_format: date` — i.e. it exercises
  exactly the repeat-sync path this gap affects, not the fresh-sync path. `eod`/`splits`/`dividends`
  each therefore carry a `streams.json` stream-level `"conformance": {"skip_dynamic": true, "reason":
  "..."}` marker (`docs/migration/conventions.md` §4) rather than a fixture workaround: rewriting the
  fixture's `date` value to a colon-offset RFC3339 string would make the check pass but would
  misrepresent Marketstack's real wire shape (the exact "recorded-real-shape" violation §4 forbids).
  No `paritytest/marketstack` exists to cite as a live substitute (this bundle has no hooks); the
  marker's reason is the honest ENGINE_GAP description above, not a claim of alternate proof.
- `page_size`/`max_pages` are not exposed as config (see Streams notes above) — only the static,
  bundle-declared page size (100, matching legacy's default) governs every request's `limit` and the
  pagination stop threshold. This never changes emitted record DATA for any input legacy itself
  would accept; it narrows configurability only.
- The full Marketstack API surface (intraday prices, standalone currencies/timezones lookup
  endpoints) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
