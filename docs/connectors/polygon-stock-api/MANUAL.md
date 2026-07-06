# pm connectors inspect polygon-stock-api

```text
NAME
  pm connectors inspect polygon-stock-api - Polygon Stock API connector manual

SYNOPSIS
  pm connectors inspect polygon-stock-api
  pm connectors inspect polygon-stock-api --json
  pm credentials add <name> --connector polygon-stock-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Polygon.io stock tickers, dividends, and splits through the Polygon.io reference REST API.

ICON
  asset: icons/polygon.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://polygon.io/docs/stocks/getting-started

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  active
  base_url
  ex_dividend_date
  execution_date
  locale
  market
  mode
  order
  page_size
  sort
  ticker
  type
  api_key (secret)

ETL STREAMS
  tickers:
    primary key: ticker
    fields: active(), currency_name(), locale(), market(), name(), primary_exchange(), ticker()
  dividends:
    primary key: id
    cursor: ex_dividend_date
    fields: cash_amount(), currency(), ex_dividend_date(), id(), ticker()
  splits:
    primary key: id
    cursor: execution_date
    fields: execution_date(), id(), split_from(), split_to(), ticker()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Polygon.io API read of stock reference data (tickers, dividends, splits)
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect polygon-stock-api

  # Inspect as structured JSON
  pm connectors inspect polygon-stock-api --json

AGENT WORKFLOW
  - Run pm connectors inspect polygon-stock-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
