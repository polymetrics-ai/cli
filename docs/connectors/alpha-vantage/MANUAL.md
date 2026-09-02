# pm connectors inspect alpha-vantage

```text
NAME
  pm connectors inspect alpha-vantage - Alpha Vantage connector manual

SYNOPSIS
  pm connectors inspect alpha-vantage
  pm connectors inspect alpha-vantage --json
  pm credentials add <name> --connector alpha-vantage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads daily, weekly, monthly, and intraday OHLCV time series plus latest global quotes through fixed Alpha Vantage query operations.

ICON
  id: alpha-vantage
  asset: icons/alpha-vantage.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  adjusted
  interval
  outputsize
  symbol (required)
  api_key (secret) (required)

ETL STREAMS
  time_series_daily:
    primary key: symbol, date
    cursor: date
    fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
  time_series_weekly:
    primary key: symbol, date
    cursor: date
    fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
  time_series_monthly:
    primary key: symbol, date
    cursor: date
    fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
  time_series_intraday:
    primary key: symbol, date
    cursor: date
    fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
  global_quote:
    primary key: symbol, latest_trading_day
    cursor: latest_trading_day
    fields: change(number), change_percent(string), high(number), latest_trading_day(string), low(number), open(number), previous_close(number), price(number), symbol(string), volume(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Alpha Vantage query reads use the fixed provider origin and declared API-key query authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect alpha-vantage

  # Inspect as structured JSON
  pm connectors inspect alpha-vantage --json

AGENT WORKFLOW
  - Run pm connectors inspect alpha-vantage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
