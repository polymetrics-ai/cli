# pm connectors inspect yahoo-finance-price

```text
NAME
  pm connectors inspect yahoo-finance-price - Yahoo Finance Price connector manual

SYNOPSIS
  pm connectors inspect yahoo-finance-price
  pm connectors inspect yahoo-finance-price --json
  pm credentials add <name> --connector yahoo-finance-price [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public Yahoo Finance chart prices as declaration-bound OHLCV records. Read-only.

ICON
  id: yahoo-finance-price
  asset: icons/yahoo-finance-price.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.yahoofinanceapi.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  interval
  range
  symbol

ETL STREAMS
  prices:
    primary key: symbol, timestamp
    cursor: timestamp
    fields: adjclose(number), close(number), currency(string), high(number), low(number), open(number), symbol(string), timestamp(number), volume(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Yahoo Finance chart API reads through a fixed declared route
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect yahoo-finance-price

  # Inspect as structured JSON
  pm connectors inspect yahoo-finance-price --json

AGENT WORKFLOW
  - Run pm connectors inspect yahoo-finance-price before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
