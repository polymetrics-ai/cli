# pm connectors inspect finnworlds

```text
NAME
  pm connectors inspect finnworlds - Finnworlds connector manual

SYNOPSIS
  pm connectors inspect finnworlds
  pm connectors inspect finnworlds --json
  pm credentials add <name> --connector finnworlds [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads global financial data (dividends, stock splits, historical candlesticks, and commodity prices) from the Finnworlds REST API.

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
  commodities
  tickers
  key (secret)

ETL STREAMS
  dividends:
    primary key: ticker, date
    cursor: date
    fields: date(), dividend_rate(), ticker()
  stock_splits:
    primary key: ticker, date
    cursor: date
    fields: date(), stock_split(), ticker()
  historical_candlestick:
    primary key: ticker, date
    cursor: date
    fields: adjusted_close(), close(), closetime(), date(), high(), low(), open(), opentime(), ticker(), trade_volume()
  commodities:
    primary key: commodity_name, datetime
    cursor: datetime
    fields: commodity_name(), commodity_price(), commodity_unit(), datetime(), percentage_day(), percentage_month(), percentage_week(), percentage_year(), price_change_day()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Finnworlds API read of global financial/market data for the configured tickers/commodities
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect finnworlds

  # Inspect as structured JSON
  pm connectors inspect finnworlds --json

AGENT WORKFLOW
  - Run pm connectors inspect finnworlds before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
