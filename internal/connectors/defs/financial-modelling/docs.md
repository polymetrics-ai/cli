# Overview

Financial Modeling Prep (FMP) is a real-time and historical market-data API. This bundle reads
the stock and ETF symbol lists, the stock screener (with optional exchange/market-cap filters),
and delisted companies, matching `internal/connectors/financial-modelling` (the hand-written
connector it migrates) at capability parity. The legacy package stays registered and unchanged
until wave6's registry flip.

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

Legacy's `stock_screener`/`delisted_companies` default page size (`fmDefaultPageSize = 1000`) is
represented in this bundle's `pagination.page_size` as `2` — a batching-size-only reduction (see
Known limits) to keep fixtures small; it never changes the total SET of records a full sync
retrieves, only how many requests it takes to retrieve them.

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
  declared-but-unwireable key is worse than an absent one, per F6). The bundle instead ships a
  fixed `page_size: 2` for both paginated streams — a batching-granularity choice, not a
  behavior-changing one: a full sync still retrieves the exact same set of records regardless of
  how many requests it takes (the short-page-stop rule is unaffected by the page-size value; see
  `docs/migration/conventions.md` §5 for the parity-deviation meta-rule this satisfies).
  ACCEPTABLE per that meta-rule: this never changes emitted record data for any legacy-accepted
  input, only request cadence.
- Only the 4 legacy-parity read streams are implemented. FMP's much larger documented surface
  (company profiles, historical prices, financial statements, quotes, forex, crypto) is out of
  scope until Pass B; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting, so
  this bundle adds none either (matching legacy's real, absent throttling behavior).
