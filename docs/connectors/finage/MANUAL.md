# pm connectors inspect finage

```text
NAME
  pm connectors inspect finage - Finage connector manual

SYNOPSIS
  pm connectors inspect finage
  pm connectors inspect finage --json
  pm credentials add <name> --connector finage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Finage US market data: most active stocks, top gainers and losers, sector performance, delisted companies, and per-symbol market news via the Finage REST API.

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
  calendar_from
  calendar_to
  mode
  symbols
  api_key (secret)

ETL STREAMS
  most_active_us_stocks:
    primary key: symbol
    fields: change(), change_percentage(), company_name(), price(), symbol()
  most_gainers:
    primary key: symbol
    fields: change(), change_percentage(), company_name(), price(), symbol()
  most_losers:
    primary key: symbol
    fields: change(), change_percentage(), company_name(), price(), symbol()
  sector_performance:
    primary key: sector
    fields: change_percentage(), sector()
  delisted_companies:
    primary key: symbol
    fields: company_name(), delisted_date(), exchange(), ipo_date(), symbol()
  market_news:
    primary key: url
    fields: date(), description(), source(), symbol(), title(), url()
  earnings_calendar:
    primary key: symbol, date
    fields: date(), eps(), estimated_eps(), estimated_revenue(), revenue(), symbol(), time()
  ipo_calendar:
    primary key: symbol, date
    fields: company(), date(), exchange(), market_cap(), price_range(), shares(), status(), symbol()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Finage API read of market data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect finage

  # Inspect as structured JSON
  pm connectors inspect finage --json

AGENT WORKFLOW
  - Run pm connectors inspect finage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
