# Overview

StockData reads tickers, end-of-day prices, intraday prices, and market news through the
stockdata.org REST API (`https://api.stockdata.org/v1`). This bundle migrates
`internal/connectors/stockdata` (the hand-written legacy connector) at capability parity; the
legacy package stays registered and unchanged until wave6's registry flip. StockData is read-only
here — legacy has no write surface, so `capabilities.write` is `false` and no `writes.json` is
shipped.

## Auth setup

Provide a StockData API access token via the `api_token` secret. It is sent as the `api_token`
query parameter on every request (`mode: api_key_query`), matching legacy's
`connsdk.APIKeyQuery("api_token", token)` exactly; it is never logged.

## Streams notes

All 4 streams (`tickers`, `eod_prices`, `intraday_prices`, `news`) share StockData's page-number
pagination (`pagination.type: page_number`, `page_param: page`, `size_param: limit`, `start_page:
1`), records at `data`. Default page size is 100 (matches legacy's `defaultPageSize`); the
`tickers` stream declares a stream-level `pagination.page_size: 2` override purely so its 2-page
conformance fixture can exercise pagination termination without needing 101 synthetic records — the
live default of 100 is unaffected in production reads (stream-level `pagination` replaces the
base-level spec wholesale only for streams that declare their own block).

`eod_prices` and `intraday_prices` both require the `symbols` config value (a plain, non-optional
`{{ config.symbols }}` query template hard-errors if absent, exactly matching legacy's
`stockdata stream requires config symbols` error for these two streams only); `date_from`/`date_to`
are optional per-request date-range filters (`omit_when_absent: true`), sent only when configured.
`tickers` and `news` never reference `symbols`/`date_from`/`date_to` in their `query` blocks, so
those streams read successfully with none of them set — matching legacy's `needsSymbols: false`
routing for both.

Legacy's published stream catalog declares `CursorFields: ["date"]` for `eod_prices`/
`intraday_prices` (this bundle's schemas mirror that with `x-cursor-field: date`), but legacy never
actually derives a request filter from a persisted incremental cursor for either stream — the only
date-range filtering is the static, config-driven `date_from`/`date_to` pair read directly from
`RuntimeConfig.Config`, never from `req.State`. This bundle reproduces that exactly: neither stream
declares a streams.json `incremental` block, so both are full-refresh reads (config-driven date
range only, no cursor-driven repeat-sync narrowing) — declaring `client_filtered`/`incremental` here
would silently drop records on a repeat sync that legacy would re-emit, an unacceptable behavior
change.

## Write actions & risks

None. StockData is read-only in legacy (`Capabilities.Write` is `false`); `Write` always returns
`connectors.ErrUnsupportedOperation`. No `writes.json` is shipped for this bundle.

## Known limits

- Full StockData API surface (real-time quotes, market status, dividends, splits) is out of scope
  for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `eod_prices`/`intraday_prices`/`news` are full-refresh only in practice (see Streams notes above)
  even though their schemas declare `x-cursor-field` for parity with legacy's advertised catalog
  fields — no `incremental` block is wired for any of them, matching legacy's actual (non-)behavior.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  StockData, so none is added here (matching legacy's real, lack-of, throttling behavior).
