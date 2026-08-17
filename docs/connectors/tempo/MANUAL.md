# pm connectors inspect tempo

```text
NAME
  pm connectors inspect tempo - Tempo connector manual

SYNOPSIS
  pm connectors inspect tempo
  pm connectors inspect tempo --json
  pm credentials add <name> --connector tempo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Tempo accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4.

ICON
  asset: icons/tempo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.tempo.io/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_token (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: global(), id(), key(), monthly_budget(), name(), status()
  customers:
    primary key: id
    fields: id(), key(), name()
  worklogs:
    primary key: tempo_worklog_id
    cursor: updated_at
    fields: billable_seconds(), created_at(), description(), issue_id(), jira_worklog_id(), start_date(), start_time(), tempo_worklog_id(), time_spent_seconds(), updated_at()
  workload_schemes:
    primary key: id
    fields: default_scheme(), description(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Tempo Cloud API read of account, customer, and worklog data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tempo

  # Inspect as structured JSON
  pm connectors inspect tempo --json

AGENT WORKFLOW
  - Run pm connectors inspect tempo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
