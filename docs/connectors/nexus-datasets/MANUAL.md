# pm connectors inspect nexus-datasets

```text
NAME
  pm connectors inspect nexus-datasets - Infor Nexus Datasets connector manual

SYNOPSIS
  pm connectors inspect nexus-datasets
  pm connectors inspect nexus-datasets --json
  pm credentials add <name> --connector nexus-datasets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads records from a configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1) using HMAC-SHA256 request signing. Read-only.

ICON
  asset: icons/nexus-datasets.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  dataset_name
  mode
  start_date
  access_key_id (secret)
  api_key (secret)
  secret_key (secret)
  user_id (secret)

ETL STREAMS
  datasets:
    primary key: id
    cursor: updated_at
    fields: dataset_name(), id(), raw_data(), raw_data_string(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Infor Nexus dataset export read, HMAC-signed
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nexus-datasets

  # Inspect as structured JSON
  pm connectors inspect nexus-datasets --json

AGENT WORKFLOW
  - Run pm connectors inspect nexus-datasets before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
