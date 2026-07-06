# pm connectors inspect open-exchange-rates

```text
NAME
  pm connectors inspect open-exchange-rates - Open Exchange Rates connector manual

SYNOPSIS
  pm connectors inspect open-exchange-rates
  pm connectors inspect open-exchange-rates --json
  pm credentials add <name> --connector open-exchange-rates [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Open Exchange Rates account usage/plan status through the Open Exchange Rates JSON API (read-only). Live/historical/currencies rate-map streams remain quarantined (ENGINE_GAP).

ICON
  asset: icons/open-exchange-rates.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.openexchangerates.org/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  app_id (secret)

ETL STREAMS
  usage:
    primary key: app_id
    fields: app_id(), daily_average(), days_elapsed(), days_remaining(), plan(), requests(), requests_quota(), requests_remaining(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Open Exchange Rates API read of account usage/plan status
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect open-exchange-rates

  # Inspect as structured JSON
  pm connectors inspect open-exchange-rates --json

AGENT WORKFLOW
  - Run pm connectors inspect open-exchange-rates before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
