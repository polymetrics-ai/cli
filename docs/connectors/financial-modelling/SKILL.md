---
name: pm-financial-modelling
description: Financial Modelling connector knowledge and safe action guide.
---

# pm-financial-modelling

## Purpose

Reads stock and ETF symbol lists, the stock screener, delisted companies, market indexes, S&P 500 constituents, the earnings calendar, and per-symbol company profiles, quotes, historical prices, financial statements, key metrics, and ratios from the Financial Modeling Prep REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- exchange
- marketcaplowerthan
- marketcapmorethan
- mode
- symbols
- api_key (secret)

## ETL Streams

- stock_screener:
  - primary key: symbol
  - fields: beta(), company_name(), country(), exchange(), exchange_short_name(), industry(), is_actively_trading(), is_etf(), last_annual_dividend(), market_cap(), price(), sector(), symbol(), volume()
- delisted_companies:
  - primary key: symbol
  - fields: company_name(), delisted_date(), exchange(), ipo_date(), symbol()
- stocks:
  - primary key: symbol
  - fields: exchange(), exchange_short_name(), name(), price(), symbol(), type()
- etfs:
  - primary key: symbol
  - fields: exchange(), exchange_short_name(), name(), price(), symbol(), type()
- market_indexes:
  - primary key: symbol
  - fields: avg_volume(), change(), changes_percentage(), day_high(), day_low(), name(), open(), previous_close(), price(), price_avg200(), price_avg50(), symbol(), timestamp(), volume(), year_high(), year_low()
- sp500_constituent:
  - primary key: symbol
  - fields: cik(), date_first_added(), founded(), head_quarter(), name(), sector(), sub_sector(), symbol()
- earnings_calendar:
  - primary key: symbol, date
  - cursor: date
  - fields: date(), eps(), eps_estimated(), fiscal_date_ending(), revenue(), revenue_estimated(), symbol(), time(), updated_from_date()
- company_profile:
  - primary key: symbol
  - fields: beta(), ceo(), company_name(), country(), currency(), description(), exchange(), exchange_short_name(), full_time_employees(), image(), industry(), ipo_date(), is_actively_trading(), is_adr(), is_etf(), is_fund(), last_div(), market_cap(), price(), sector(), symbol(), vol_avg(), website()
- quote:
  - primary key: symbol
  - fields: avg_volume(), change(), changes_percentage(), day_high(), day_low(), earnings_announcement(), eps(), exchange(), market_cap(), name(), open(), pe(), previous_close(), price(), price_avg200(), price_avg50(), shares_outstanding(), symbol(), timestamp(), volume(), year_high(), year_low()
- historical_price:
  - primary key: symbol, date
  - cursor: date
  - fields: adj_close(), change(), change_percent(), close(), date(), high(), low(), open(), symbol(), volume(), vwap()
- income_statement:
  - primary key: symbol, date, period
  - cursor: date
  - fields: accepted_date(), calendar_year(), cost_of_revenue(), date(), eps(), eps_diluted(), filling_date(), fiscal_year(), gross_profit(), gross_profit_ratio(), net_income(), net_income_ratio(), operating_expenses(), operating_income(), operating_income_ratio(), period(), reported_currency(), revenue(), symbol()
- balance_sheet_statement:
  - primary key: symbol, date, period
  - cursor: date
  - fields: accepted_date(), calendar_year(), cash_and_cash_equivalents(), date(), filling_date(), fiscal_year(), period(), reported_currency(), symbol(), total_assets(), total_current_assets(), total_current_liabilities(), total_equity(), total_liabilities(), total_liabilities_and_total_equity()
- cash_flow_statement:
  - primary key: symbol, date, period
  - cursor: date
  - fields: accepted_date(), calendar_year(), capital_expenditure(), date(), filling_date(), fiscal_year(), free_cash_flow(), net_cash_provided_by_operating_activities(), net_change_in_cash(), net_income(), operating_cash_flow(), period(), reported_currency(), symbol()
- key_metrics:
  - primary key: symbol, date, period
  - cursor: date
  - fields: calendar_year(), current_ratio(), date(), debt_to_equity(), enterprise_value(), free_cash_flow_yield(), market_cap(), pb_ratio(), pe_ratio(), period(), revenue_per_share(), roe(), symbol()
- financial_ratios:
  - primary key: symbol, date, period
  - cursor: date
  - fields: calendar_year(), current_ratio(), date(), debt_ratio(), dividend_yield(), gross_profit_margin(), net_profit_margin(), period(), price_earnings_ratio(), quick_ratio(), return_on_assets(), return_on_equity(), symbol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Financial Modeling Prep API read of market data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect financial-modelling
```

### Inspect as structured JSON

```bash
pm connectors inspect financial-modelling --json
```

## Agent Rules

- Run pm connectors inspect financial-modelling before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
