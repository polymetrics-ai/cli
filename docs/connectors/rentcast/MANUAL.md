# pm connectors inspect rentcast

```text
NAME
  pm connectors inspect rentcast - RentCast connector manual

SYNOPSIS
  pm connectors inspect rentcast
  pm connectors inspect rentcast --json
  pm credentials add <name> --connector rentcast [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads RentCast properties, sale listings, rental listings, market data, and value/rental estimates through the RentCast REST API. Read-only.

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
  address
  base_url
  city
  property_type
  state
  zip_code
  api_key (secret)

ETL STREAMS
  properties:
    primary key: id
    cursor: last_seen_date
    fields: address(), city(), id(), last_seen_date(), property_type(), state(), zip_code()
  sale_listings:
    primary key: id
    cursor: last_seen_date
    fields: address(), id(), last_seen_date(), price(), property_type()
  rental_listings:
    primary key: id
    cursor: last_seen_date
    fields: address(), id(), last_seen_date(), property_type(), rent()
  markets:
    primary key: id
    fields: city(), id(), state(), zip_code()
  value_estimates:
    primary key: id
    fields: address(), id(), price()
  rental_estimates:
    primary key: id
    fields: address(), id(), rent()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external RentCast API read of property, listing, market, and valuation data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rentcast

  # Inspect as structured JSON
  pm connectors inspect rentcast --json

AGENT WORKFLOW
  - Run pm connectors inspect rentcast before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
