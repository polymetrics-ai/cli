# pm connectors inspect sample

```text
NAME
  pm connectors inspect sample - Sample connector manual

SYNOPSIS
  pm connectors inspect sample
  pm connectors inspect sample --json
  pm credentials add <name> --connector sample [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Built-in deterministic source connector for local development and tests.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

ETL STREAMS
  customers: Sample customer records.
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), email(string), plan(string), updated_at(timestamp)
  events: Sample event records.
    primary key: id
    cursor: occurred_at
    fields: id(string), customer_id(string), event(string), occurred_at(timestamp)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
  Source modes: full_refresh, incremental
  Destination modes: append, overwrite, append_dedup, overwrite_dedup

SECURITY
  read risk: local deterministic sample data
  write risk: unsupported
  mutation risk: none
  approval: not required for reads
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sample

  # Inspect as structured JSON
  pm connectors inspect sample --json

  # Sample ETL
  pm credentials add sample-local --connector sample
  pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers
  pm etl run --connection sample_to_warehouse --stream customers --json

AGENT WORKFLOW
  - Run pm connectors inspect sample before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
