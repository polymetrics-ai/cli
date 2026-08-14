# pm connectors inspect plausible

```text
NAME
  pm connectors inspect plausible - Plausible connector manual

SYNOPSIS
  pm connectors inspect plausible
  pm connectors inspect plausible --json
  pm credentials add <name> --connector plausible [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Plausible Analytics sites and stats reports through the Stats API.

ICON
  id: plausible
  asset: icons/plausible.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://plausible.io/docs/stats-api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  compare
  date
  filters
  metrics
  mode
  period
  property
  site_id
  api_token (secret) (required)

ETL STREAMS
  sites:
    primary key: site_id
    fields: domain(string), site_id(string)
  aggregate:
    primary key: site_id
    fields: bounce_rate(number), events(integer), pageviews(integer), site_id(string), visit_duration(number), visitors(integer), visits(integer)
  timeseries:
    primary key: date
    fields: bounce_rate(number), date(string), events(integer), pageviews(integer), site_id(string), visit_duration(number), visitors(integer), visits(integer)
  breakdown:
    primary key: property_value
    fields: bounce_rate(number), events(integer), pageviews(integer), property_value(string), site_id(string), visit_duration(number), visitors(integer), visits(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Plausible Analytics API read of site analytics data
  approval: none; read-only analytics sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect plausible

  # Inspect as structured JSON
  pm connectors inspect plausible --json

AGENT WORKFLOW
  - Run pm connectors inspect plausible before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
