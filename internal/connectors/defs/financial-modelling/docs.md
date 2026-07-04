# Overview

Financial Modeling Prep (FMP) is a real-time and historical market-data API. This bundle originally
migrated `internal/connectors/financial-modelling` (the hand-written connector, still registered
and unchanged until wave6's registry flip) at capability parity — the stock and ETF symbol lists,
the stock screener, and delisted companies. This Pass B pass expands to FMP's much wider documented
surface: 11 aggregate/list streams (`stocks`, `etfs`, `stock_screener`, `delisted_companies`,
`market_indexes`, `sp500_constituent`, `earnings_calendar`) plus 8 new per-symbol fan-out streams
(`company_profile`, `quote`, `historical_price`, `income_statement`, `balance_sheet_statement`,
`cash_flow_statement`, `key_metrics`, `financial_ratios`) driven by a new `symbols` config value.
FMP is a read-only market-data API — no write/mutation endpoint exists anywhere in its documented
surface — so `capabilities.write` stays `false` and no `writes.json` is declared.

## Auth setup

Provide an FMP API key via the `api_key` secret; it is sent only as the `apikey` query parameter
(`auth: [{"mode": "api_key_query", "param": "apikey", ...}]`) and is never logged. No other
credential shape exists for this API.

## Streams notes

`stocks` (`GET /stock/list`) and `etfs` (`GET /etf/list`) are single-request, unpaginated list
endpoints — the full symbol universe returns in one response, matching legacy's non-paginated
`harvest` branch. Both rename the raw `exchangeShortName` field to `exchange_short_name` via
`computed_fields`.

`stock_screener` (`GET /stock-screener`) and `delisted_companies` (`GET /delisted-companies`)
honor limit/offset pagination (`pagination.type: offset_limit`), matching legacy's paginated
`harvest` branch: a full page implies another page exists, a short (or empty) page ends the read.
`stock_screener` additionally accepts three optional screener filters — `exchange`,
`marketcapmorethan` (sent as `marketCapMoreThan`), `marketcaplowerthan` (sent as
`marketCapLowerThan`) — via the optional-query dialect (`omit_when_absent: true`), matching
legacy's `fmScreenerParams` (each filter is sent only when its config value is set; absent by
default). Both paginated streams' record mappers use `computed_fields` renames
(`companyName`→`company_name`, `marketCap`→`market_cap`, `lastAnnualDividend`→
`last_annual_dividend`, `exchangeShortName`→`exchange_short_name`, `isEtf`→`is_etf`,
`isActivelyTrading`→`is_actively_trading`, `ipoDate`→`ipo_date`, `delistedDate`→`delisted_date`).
Numeric/boolean fields (`market_cap`, `beta`, `volume`, `is_etf`, `is_actively_trading`) use bare
`{{ record.<path> }}` computed_fields references, which the engine's typed-extraction rule copies
as their native JSON type (not stringified).

`stock_screener`'s `pagination.page_size` is `1000`, matching legacy's real production default
(`fmDefaultPageSize = 1000`) — this is the actual value a live deployment's paginator sends, not a
fixture convenience (see Known limits for why legacy's runtime override isn't wired).
`delisted_companies` declares a stream-level `pagination` override (`page_size: 2`) so its required
2-page conformance fixture (`fixtures/streams/delisted_companies/{page_1,page_2}.json`, §4 of
`docs/migration/conventions.md`) can stay small and readable; since stream-level `pagination`
replaces the base spec wholesale, this is an intentional, ledgered per-stream deviation from
legacy's uniform 1000-record page size — `delisted_companies` reads in smaller, more numerous pages
than legacy would, `stock_screener` is unaffected and uses legacy's true 1000-record page size
end-to-end (matching its fixture's `limit=1000` request/response). Neither deviation changes the
total SET of records a full sync retrieves, only how many requests it takes to retrieve them.

**New in this pass — aggregate/list streams, no fan-out:**

- **`market_indexes`** (`GET /quotes/index`): unpaginated list of major index quotes (S&P 500,
  Dow, Nasdaq, etc). `computed_fields` renames the raw camelCase fields
  (`changesPercentage`/`dayLow`/`dayHigh`/`yearHigh`/`yearLow`/`priceAvg50`/`priceAvg200`/
  `avgVolume`/`previousClose`) to their snake_case schema names.
- **`sp500_constituent`** (`GET /sp500_constituent`): current S&P 500 membership list.
  `computed_fields` renames `subSector`/`headQuarter`/`dateFirstAdded`/`founded`.
- **`earnings_calendar`** (`GET /earning_calendar`): upcoming earnings announcement dates across
  all symbols (unfiltered — no `from`/`to` window is wired; see Known limits). Primary key
  `[symbol, date]`, cursor field `date`. `computed_fields` renames `epsEstimated`/
  `revenueEstimated`/`updatedFromDate`/`fiscalDateEnding`.

**New in this pass — per-symbol fan-out streams**, driven by the new `symbols` spec property (a
comma-separated ticker list, matching this bundle's existing `finnhub`/`finnworlds` sibling
connectors' `symbols`/`tickers` config shape): each declares `fan_out.ids_from.config_key: symbols`
and `into.path_var`, since every one of these FMP endpoints takes the symbol as a path segment
(`/profile/{symbol}`, not a query parameter) — `stamp_field: symbol` (or `ticker` for
`company_profile`, matching this bundle's existing convention of stamping the fan-out id onto a
per-stream-appropriate field name) writes the requested symbol onto every emitted record. None of
these 8 endpoints paginate — each returns its full per-symbol result (a 1-element array for the
snapshot-shaped endpoints, or a `historical`-nested array for the time-series endpoint) in a single
response, so no stream declares a `pagination` block.

- **`company_profile`** (`GET /profile/{symbol}`): full company profile (sector, industry,
  description, CEO, market cap, etc). `computed_fields` renames `companyName`/
  `exchangeShortName`/`mktCap`(→`market_cap`)/`lastDiv`/`isEtf`/`isActivelyTrading`/`isAdr`/
  `isFund`/`fullTimeEmployees`/`ipoDate`.
- **`quote`** (`GET /quote/{symbol}`): real-time quote snapshot (price, day/year range, volume,
  EPS/PE, earnings announcement date). `computed_fields` renames the same camelCase-to-snake_case
  shape as `market_indexes`, plus `earningsAnnouncement`/`sharesOutstanding`.
- **`historical_price`** (`GET /historical-price-full/{symbol}`): daily OHLCV history; records
  live at `historical` (the response wraps the array under `{"symbol": ..., "historical": [...]}`
  rather than at the response root). Primary key `[symbol, date]`, cursor field `date`.
  `computed_fields` renames `adjClose`/`changePercent`.
- **`income_statement`** / **`balance_sheet_statement`** / **`cash_flow_statement`**
  (`GET /income-statement/{symbol}`, `/balance-sheet-statement/{symbol}`,
  `/cash-flow-statement/{symbol}`): the 3 core annual financial statements (FMP's default
  `period` is annual; no `period=quarter` override is wired — see Known limits). Primary key
  `[symbol, date, period]`, cursor field `date`; each renames its own set of camelCase report
  fields (`reportedCurrency`/`fiscalYear`/`fillingDate`/`acceptedDate`/`calendarYear` are common to
  all three, plus statement-specific fields — `grossProfit`/`netIncome`/`operatingIncome` for
  income, `totalAssets`/`totalLiabilities`/`totalEquity` for balance sheet,
  `operatingCashFlow`/`freeCashFlow`/`capitalExpenditure` for cash flow).
- **`key_metrics`** (`GET /key-metrics/{symbol}`) and **`financial_ratios`**
  (`GET /ratios/{symbol}`): derived valuation/profitability/liquidity metrics computed by FMP
  from the underlying statements (market cap, PE/PB ratios, current/quick ratios, ROE, free cash
  flow yield, etc). Same `[symbol, date, period]` primary key / `date` cursor shape as the 3
  statement streams.

## Auth setup

Provide an FMP API key via the `api_key` secret; it is sent only as the `apikey` query parameter
(`auth: [{"mode": "api_key_query", "param": "apikey", ...}]`) and is never logged. No other
credential shape exists for this API.

## Streams notes

`stocks` (`GET /stock/list`) and `etfs` (`GET /etf/list`) are single-request, unpaginated list
endpoints — the full symbol universe returns in one response, matching legacy's non-paginated
`harvest` branch. Both rename the raw `exchangeShortName` field to `exchange_short_name` via
`computed_fields`.

`stock_screener` (`GET /stock-screener`) and `delisted_companies` (`GET /delisted-companies`)
honor limit/offset pagination (`pagination.type: offset_limit`), matching legacy's paginated
`harvest` branch: a full page implies another page exists, a short (or empty) page ends the read.
`stock_screener` additionally accepts three optional screener filters — `exchange`,
`marketcapmorethan` (sent as `marketCapMoreThan`), `marketcaplowerthan` (sent as
`marketCapLowerThan`) — via the optional-query dialect (`omit_when_absent: true`), matching
legacy's `fmScreenerParams` (each filter is sent only when its config value is set; absent by
default). Both paginated streams' record mappers use `computed_fields` renames
(`companyName`→`company_name`, `marketCap`→`market_cap`, `lastAnnualDividend`→
`last_annual_dividend`, `exchangeShortName`→`exchange_short_name`, `isEtf`→`is_etf`,
`isActivelyTrading`→`is_actively_trading`, `ipoDate`→`ipo_date`, `delistedDate`→`delisted_date`).
Numeric/boolean fields (`market_cap`, `beta`, `volume`, `is_etf`, `is_actively_trading`) use bare
`{{ record.<path> }}` computed_fields references, which the engine's typed-extraction rule copies
as their native JSON type (not stringified).

`stock_screener`'s `pagination.page_size` is `1000`, matching legacy's real production default
(`fmDefaultPageSize = 1000`) — this is the actual value a live deployment's paginator sends, not a
fixture convenience (see Known limits for why legacy's runtime override isn't wired).
`delisted_companies` declares a stream-level `pagination` override (`page_size: 2`) so its required
2-page conformance fixture (`fixtures/streams/delisted_companies/{page_1,page_2}.json`, §4 of
`docs/migration/conventions.md`) can stay small and readable; since stream-level `pagination`
replaces the base spec wholesale, this is an intentional, ledgered per-stream deviation from
legacy's uniform 1000-record page size — `delisted_companies` reads in smaller, more numerous pages
than legacy would, `stock_screener` is unaffected and uses legacy's true 1000-record page size
end-to-end (matching its fixture's `limit=1000` request/response). Neither deviation changes the
total SET of records a full sync retrieves, only how many requests it takes to retrieve them.

## Write actions & risks

None. Financial Modeling Prep is a read-only market-data source in this bundle; `capabilities.write`
is `false` and no `writes.json` is declared, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`).

## Known limits

- Legacy's runtime-configurable `page_size` (`fmPageSize` config key, default 1000, max 10000) and
  `max_pages` (`fmMaxPages` config key) are **not wired** in this bundle: the engine's
  `PaginationSpec.PageSize`/`MaxPages` fields are static JSON integers on `streams.json`, not
  templated against `config.*` — there is no mechanism to make either config-overridable in the
  declarative dialect (same class as searxng's dead `page_size`/`max_pages` config, §1 of
  `docs/migration/conventions.md`). Both config keys are therefore not declared in `spec.json` (a
  declared-but-unwireable key is worse than an absent one, per F6). `stock_screener`'s
  `pagination.page_size` is fixed at legacy's own default (`1000`), reproducing legacy's
  default-configuration behavior exactly; `delisted_companies` keeps a smaller `page_size: 2` as a
  ledgered per-stream deviation purely to keep its 2-page conformance fixture small (see Streams
  notes) — an operator who had overridden legacy's `page_size` away from its default cannot
  reproduce that override here, but every request this bundle sends by default matches legacy's own
  default cadence. Neither deviation changes the total SET of records a full sync retrieves, only
  how many requests it takes to retrieve them (the short-page-stop rule is unaffected by the
  page-size value; see `docs/migration/conventions.md` §5 for the parity-deviation meta-rule this
  satisfies). ACCEPTABLE per that meta-rule: this never changes emitted record data for any
  legacy-accepted input, only request cadence.
- **8 per-symbol streams require `symbols` to be configured** (no auto-discovery fallback — an
  operator who leaves `symbols` unset gets zero fan-out ids and therefore zero records from
  `company_profile`/`quote`/`historical_price`/`income_statement`/`balance_sheet_statement`/
  `cash_flow_statement`/`key_metrics`/`financial_ratios`, matching `stream.fan_out`'s documented
  `config_key` semantics — an empty/absent CSV yields a nil id list, not an error). `symbols` is
  not referenced by the 7 aggregate/list streams.
- **`period=quarter` is not wired** for `income_statement`/`balance_sheet_statement`/
  `cash_flow_statement`/`key_metrics`/`financial_ratios`: FMP's real default (no `period` query
  param sent) is annual, which is exactly what this bundle sends — an operator who wants quarterly
  granularity has no config override here. This narrows FMP's documented surface (see
  `api_surface.json`'s `duplicate_of` entries for the `period=quarter` variants) but never changes
  the DATA an operator relying on the (annual) default would see.
- **`historical_price`'s `from`/`to` date-window query params are not wired**: FMP's
  `historical-price-full` endpoint accepts optional `from`/`to` bounds to narrow the returned
  range, but every request this bundle sends is unfiltered (FMP's own full-history default when
  neither is supplied) — an operator relying on the endpoint's default range sees identical data;
  an operator who wanted a narrower explicit window cannot express it here. No `incremental` block
  is declared on this stream for the same reason a request-time filter isn't wired: FMP's `date`
  cursor could support `incremental.request_param` in a future pass, but this expansion prioritized
  covering more distinct FMP resources over deepening incremental support on any one of them.
- **`earnings_calendar`'s `from`/`to` date-window is not wired**, for the identical reason as
  `historical_price` above: every read is FMP's unfiltered/default-windowed response.
- FMP's much larger documented surface beyond the 15 streams above (100+ endpoints total) remains
  out of scope — see `api_surface.json` for the full per-endpoint breakdown: intraday/technical-
  indicator interval variants (a repeating 6-interval shape per asset class), bulk/CSV exports,
  SEC filing full-text/transcripts, insider trading and 13F institutional holdings, ETF/mutual-fund
  sub-holdings, forex/crypto/commodity asset classes, and regional exchange mirrors (EuroNext/TSX)
  are each excluded with a specific category+reason rather than a blanket bucket.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
