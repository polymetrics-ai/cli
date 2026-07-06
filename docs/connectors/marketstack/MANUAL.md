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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  symbols
  api_key (secret)

ETL STREAMS
  exchanges:
    primary key: mic
    fields: acronym(), city(), country(), country_code(), currency_code(), currency_name(), currency_symbol(), mic(), name(), timezone(), timezone_abbr(), website()
  tickers:
    primary key: symbol
    fields: has_eod(), has_intraday(), name(), stock_exchange_mic(), stock_exchange_name(), symbol()
  eod:
    primary key: symbol, date
    cursor: date
    fields: adj_close(), adj_high(), adj_low(), adj_open(), adj_volume(), close(), date(), dividend(), exchange(), high(), low(), open(), split_factor(), symbol(), volume()
  splits:
    primary key: symbol, date
    cursor: date
    fields: date(), split_factor(), symbol()
  dividends:
    primary key: symbol, date
    cursor: date
    fields: date(), dividend(), symbol()

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
