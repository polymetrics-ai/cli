# pm connectors inspect marketstack

```text
NAME
  pm connectors inspect marketstack - Marketstack connector manual

SYNOPSIS
  pm connectors inspect marketstack
  pm connectors inspect marketstack --json
  pm credentials add <name> --connector marketstack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Marketstack exchanges, tickers, end-of-day prices, splits, and dividends through the Marketstack REST API.

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
  start_date
  symbols
  api_key (secret) (required)

ETL STREAMS
  exchanges:
    primary key: mic
    fields: acronym(string), city(string), country(string), country_code(string), currency_code(string), currency_name(string), currency_symbol(string), mic(string), name(string), timezone(string), timezone_abbr(string), website(string)
  tickers:
    primary key: symbol
    fields: has_eod(boolean), has_intraday(boolean), name(string), stock_exchange_mic(string), stock_exchange_name(string), symbol(string)
  eod:
    primary key: symbol, date
    cursor: date
    fields: adj_close(number), adj_high(number), adj_low(number), adj_open(number), adj_volume(number), close(number), date(string), dividend(number), exchange(string), high(number), low(number), open(number), split_factor(number), symbol(string), volume(number)
  splits:
    primary key: symbol, date
    cursor: date
    fields: date(string), split_factor(number), symbol(string)
  dividends:
    primary key: symbol, date
    cursor: date
    fields: date(string), dividend(number), symbol(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Marketstack API read of financial market data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect marketstack

  # Inspect as structured JSON
  pm connectors inspect marketstack --json

AGENT WORKFLOW
  - Run pm connectors inspect marketstack before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
