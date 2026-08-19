# pm connectors inspect twelve-data

```text
NAME
  pm connectors inspect twelve-data - Twelve Data connector manual

SYNOPSIS
  pm connectors inspect twelve-data
  pm connectors inspect twelve-data --json
  pm credentials add <name> --connector twelve-data [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Twelve Data time series, quotes, stocks, and forex pair reference data.

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
  interval
  output_size
  symbol
  api_key (secret) (required)

ETL STREAMS
  time_series:
    primary key: symbol, datetime
    cursor: datetime
    fields: close(string), datetime(string), high(string), low(string), open(string), symbol(string), volume(string)
  quote:
    primary key: symbol
    fields: close(string), currency(string), name(string), symbol(string)
  stocks:
    primary key: symbol
    fields: currency(string), name(string), symbol(string)
  forex_pairs:
    primary key: symbol
    fields: currency(string), name(string), symbol(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Twelve Data API read of market time series, quote, and reference data
  approval: none; read-only, no reverse-ETL writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect twelve-data

  # Inspect as structured JSON
  pm connectors inspect twelve-data --json

AGENT WORKFLOW
  - Run pm connectors inspect twelve-data before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
