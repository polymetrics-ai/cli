# Overview

Reads stock and ETF symbol lists, the stock screener, delisted companies, market indexes, S&P 500
constituents, the earnings calendar, and per-symbol company profiles, quotes, historical prices,
financial statements, key metrics, and ratios from the Financial Modeling Prep REST API.

Readable streams: `stock_screener`, `delisted_companies`, `stocks`, `etfs`, `market_indexes`,
`sp500_constituent`, `earnings_calendar`, `company_profile`, `quote`, `historical_price`,
`income_statement`, `balance_sheet_statement`, `cash_flow_statement`, `key_metrics`,
`financial_ratios`.

This connector is read-only; no write actions are declared.

Service API documentation: https://site.financialmodelingprep.com/developer/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Financial Modeling Prep API key. Sent only as the apikey
  query parameter; never logged.
- `base_url` (optional, string); default `https://financialmodelingprep.com/api/v3`; format `uri`;
  Financial Modeling Prep API base URL override for tests or proxies.
- `exchange` (optional, string); Optional stock-screener filter: exchange name (e.g. NASDAQ).
- `marketcaplowerthan` (optional, string); Optional stock-screener filter: maximum market
  capitalization.
- `marketcapmorethan` (optional, string); Optional stock-screener filter: minimum market
  capitalization.
- `mode` (optional, string).
- `symbols` (optional, string); Comma-separated stock ticker symbols to fan out over for the
  per-symbol streams (company_profile, quote, historical_price, income_statement,
  balance_sheet_statement, cash_flow_statement, key_metrics, financial_ratios). Required for those
  streams only; the aggregate/list streams (stocks, etfs, stock_screener, delisted_companies,
  market_indexes, sp500_constituent, earnings_calendar) do not reference it.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://financialmodelingprep.com/api/v3`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/stock-screener` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `stocks`, `etfs`, `market_indexes`, `sp500_constituent`,
`earnings_calendar`, `company_profile`, `quote`, `historical_price`, `income_statement`,
`balance_sheet_statement`, `cash_flow_statement`, `key_metrics`, `financial_ratios`; offset_limit:
`stock_screener`, `delisted_companies`.

- `stock_screener`: GET `/stock-screener` - records at response root; query `exchange` from template
  `{{ config.exchange }}`, omitted when absent; `marketCapLowerThan` from template `{{
  config.marketcaplowerthan }}`, omitted when absent; `marketCapMoreThan` from template `{{
  config.marketcapmorethan }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 1000; computed output fields `company_name`,
  `exchange_short_name`, `is_actively_trading`, `is_etf`, `last_annual_dividend`, `market_cap`.
- `delisted_companies`: GET `/delisted-companies` - records at response root; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 2; computed output
  fields `company_name`, `delisted_date`, `ipo_date`.
- `stocks`: GET `/stock/list` - records at response root; computed output fields
  `exchange_short_name`.
- `etfs`: GET `/etf/list` - records at response root; computed output fields `exchange_short_name`.
- `market_indexes`: GET `/quotes/index` - records at response root; computed output fields
  `avg_volume`, `changes_percentage`, `day_high`, `day_low`, `previous_close`, `price_avg200`,
  `price_avg50`, `year_high`, `year_low`.
- `sp500_constituent`: GET `/sp500_constituent` - records at response root; computed output fields
  `date_first_added`, `founded`, `head_quarter`, `sub_sector`.
- `earnings_calendar`: GET `/earning_calendar` - records at response root; computed output fields
  `eps_estimated`, `fiscal_date_ending`, `revenue_estimated`, `updated_from_date`.
- `company_profile`: GET `/profile/{{ fanout.id }}` - records at response root; computed output
  fields `company_name`, `exchange_short_name`, `full_time_employees`, `ipo_date`,
  `is_actively_trading`, `is_adr`, `is_etf`, `is_fund`, `last_div`, `market_cap`; fan-out; ids from
  config field `symbols`; id inserted into the request path; stamps `symbol`.
- `quote`: GET `/quote/{{ fanout.id }}` - records at response root; computed output fields
  `avg_volume`, `changes_percentage`, `day_high`, `day_low`, `earnings_announcement`, `eps`,
  `market_cap`, `pe`, `previous_close`, `price_avg200`, `price_avg50`, `shares_outstanding`,
  `year_high`, `year_low`; fan-out; ids from config field `symbols`; id inserted into the request
  path; stamps `symbol`.
- `historical_price`: GET `/historical-price-full/{{ fanout.id }}` - records path `historical`;
  computed output fields `adj_close`, `change_percent`; fan-out; ids from config field `symbols`; id
  inserted into the request path; stamps `symbol`.
- `income_statement`: GET `/income-statement/{{ fanout.id }}` - records at response root; computed
  output fields `accepted_date`, `calendar_year`, `eps_diluted`, `filling_date`, `fiscal_year`,
  `gross_profit`, `gross_profit_ratio`, `net_income`, `net_income_ratio`, `operating_income`,
  `operating_income_ratio`, `reported_currency`; fan-out; ids from config field `symbols`; id
  inserted into the request path; stamps `symbol`.
- `balance_sheet_statement`: GET `/balance-sheet-statement/{{ fanout.id }}` - records at response
  root; computed output fields `accepted_date`, `calendar_year`, `cash_and_cash_equivalents`,
  `filling_date`, `fiscal_year`, `reported_currency`, `total_assets`, `total_current_assets`,
  `total_current_liabilities`, `total_equity`, `total_liabilities`; fan-out; ids from config field
  `symbols`; id inserted into the request path; stamps `symbol`.
- `cash_flow_statement`: GET `/cash-flow-statement/{{ fanout.id }}` - records at response root;
  computed output fields `accepted_date`, `calendar_year`, `capital_expenditure`, `filling_date`,
  `fiscal_year`, `free_cash_flow`, `net_cash_provided_by_operating_activities`,
  `net_change_in_cash`, `operating_cash_flow`, `reported_currency`; fan-out; ids from config field
  `symbols`; id inserted into the request path; stamps `symbol`.
- `key_metrics`: GET `/key-metrics/{{ fanout.id }}` - records at response root; computed output
  fields `calendar_year`, `current_ratio`, `debt_to_equity`, `enterprise_value`,
  `free_cash_flow_yield`, `market_cap`, `pb_ratio`, `pe_ratio`, `revenue_per_share`, `roe`; fan-out;
  ids from config field `symbols`; id inserted into the request path; stamps `symbol`.
- `financial_ratios`: GET `/ratios/{{ fanout.id }}` - records at response root; computed output
  fields `calendar_year`, `current_ratio`, `debt_ratio`, `dividend_yield`, `gross_profit_margin`,
  `net_profit_margin`, `price_earnings_ratio`, `quick_ratio`, `return_on_assets`,
  `return_on_equity`; fan-out; ids from config field `symbols`; id inserted into the request path;
  stamps `symbol`.

## Write actions & risks

This connector is read-only. Read behavior: external Financial Modeling Prep API read of market
data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 15 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=9, duplicate_of=63, non_data_endpoint=12, out_of_scope=51.
