# pm connectors inspect financial-modelling

```text
NAME
  pm connectors inspect financial-modelling - Financial Modelling connector manual

SYNOPSIS
  pm connectors inspect financial-modelling
  pm connectors inspect financial-modelling --json
  pm credentials add <name> --connector financial-modelling [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads stock and ETF symbol lists, the stock screener, delisted companies, market indexes, S&P 500 constituents, the earnings calendar, and per-symbol company profiles, quotes, historical prices, financial statements, key metrics, and ratios from the Financial Modeling Prep REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  exchange
  marketcaplowerthan
  marketcapmorethan
  mode
  symbols
  api_key (secret) (required)

ETL STREAMS
  stock_screener:
    primary key: symbol
    fields: beta(number), company_name(string), country(string), exchange(string), exchange_short_name(string), industry(string), is_actively_trading(boolean), is_etf(boolean), last_annual_dividend(number), market_cap(integer), price(number), sector(string), symbol(string), volume(integer)
  delisted_companies:
    primary key: symbol
    fields: company_name(string), delisted_date(string), exchange(string), ipo_date(string), symbol(string)
  stocks:
    primary key: symbol
    fields: exchange(string), exchange_short_name(string), name(string), price(number), symbol(string), type(string)
  etfs:
    primary key: symbol
    fields: exchange(string), exchange_short_name(string), name(string), price(number), symbol(string), type(string)
  market_indexes:
    primary key: symbol
    fields: avg_volume(integer), change(number), changes_percentage(number), day_high(number), day_low(number), name(string), open(number), previous_close(number), price(number), price_avg200(number), price_avg50(number), symbol(string), timestamp(integer), volume(integer), year_high(number), year_low(number)
  sp500_constituent:
    primary key: symbol
    fields: cik(string), date_first_added(string), founded(string), head_quarter(string), name(string), sector(string), sub_sector(string), symbol(string)
  earnings_calendar:
    primary key: symbol, date
    cursor: date
    fields: date(string), eps(number), eps_estimated(number), fiscal_date_ending(string), revenue(number), revenue_estimated(number), symbol(string), time(string), updated_from_date(string)
  company_profile:
    primary key: symbol
    fields: beta(number), ceo(string), company_name(string), country(string), currency(string), description(string), exchange(string), exchange_short_name(string), full_time_employees(string), image(string), industry(string), ipo_date(string), is_actively_trading(boolean), is_adr(boolean), is_etf(boolean), is_fund(boolean), last_div(number), market_cap(number), price(number), sector(string), symbol(string), vol_avg(integer), website(string)
  quote:
    primary key: symbol
    fields: avg_volume(integer), change(number), changes_percentage(number), day_high(number), day_low(number), earnings_announcement(string), eps(number), exchange(string), market_cap(number), name(string), open(number), pe(number), previous_close(number), price(number), price_avg200(number), price_avg50(number), shares_outstanding(number), symbol(string), timestamp(integer), volume(integer), year_high(number), year_low(number)
  historical_price:
    primary key: symbol, date
    cursor: date
    fields: adj_close(number), change(number), change_percent(number), close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer), vwap(number)
  income_statement:
    primary key: symbol, date, period
    cursor: date
    fields: accepted_date(string), calendar_year(string), cost_of_revenue(number), date(string), eps(number), eps_diluted(number), filling_date(string), fiscal_year(string), gross_profit(number), gross_profit_ratio(number), net_income(number), net_income_ratio(number), operating_expenses(number), operating_income(number), operating_income_ratio(number), period(string), reported_currency(string), revenue(number), symbol(string)
  balance_sheet_statement:
    primary key: symbol, date, period
    cursor: date
    fields: accepted_date(string), calendar_year(string), cash_and_cash_equivalents(number), date(string), filling_date(string), fiscal_year(string), period(string), reported_currency(string), symbol(string), total_assets(number), total_current_assets(number), total_current_liabilities(number), total_equity(number), total_liabilities(number), total_liabilities_and_total_equity(number)
  cash_flow_statement:
    primary key: symbol, date, period
    cursor: date
    fields: accepted_date(string), calendar_year(string), capital_expenditure(number), date(string), filling_date(string), fiscal_year(string), free_cash_flow(number), net_cash_provided_by_operating_activities(number), net_change_in_cash(number), net_income(number), operating_cash_flow(number), period(string), reported_currency(string), symbol(string)
  key_metrics:
    primary key: symbol, date, period
    cursor: date
    fields: calendar_year(string), current_ratio(number), date(string), debt_to_equity(number), enterprise_value(number), free_cash_flow_yield(number), market_cap(number), pb_ratio(number), pe_ratio(number), period(string), revenue_per_share(number), roe(number), symbol(string)
  financial_ratios:
    primary key: symbol, date, period
    cursor: date
    fields: calendar_year(string), current_ratio(number), date(string), debt_ratio(number), dividend_yield(number), gross_profit_margin(number), net_profit_margin(number), period(string), price_earnings_ratio(number), quick_ratio(number), return_on_assets(number), return_on_equity(number), symbol(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Financial Modeling Prep API read of market data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect financial-modelling

  # Inspect as structured JSON
  pm connectors inspect financial-modelling --json

AGENT WORKFLOW
  - Run pm connectors inspect financial-modelling before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
