# pm connectors inspect outbrain-amplify

```text
NAME
  pm connectors inspect outbrain-amplify - Outbrain Amplify connector manual

SYNOPSIS
  pm connectors inspect outbrain-amplify
  pm connectors inspect outbrain-amplify --json
  pm credentials add <name> --connector outbrain-amplify [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Outbrain Amplify marketers, campaigns, and performance reports via the Outbrain Amplify REST API. Read-only.

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
  conversion_count
  end_date
  geo_location_breakdown
  marketer_id
  max_pages
  mode
  page_size
  report_granularity
  start_date
  username
  access_token (secret)
  password (secret)

ETL STREAMS
  marketers:
    primary key: id
    fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)
  campaigns:
    primary key: id
    fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)
  performance_reports:
    primary key: id
    fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Outbrain Amplify API read of marketer, campaign, and performance report data
  approval: none; read-only marketing API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect outbrain-amplify

  # Inspect as structured JSON
  pm connectors inspect outbrain-amplify --json

AGENT WORKFLOW
  - Run pm connectors inspect outbrain-amplify before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
