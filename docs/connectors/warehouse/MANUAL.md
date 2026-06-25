# pm connectors inspect warehouse

```text
NAME
  pm connectors inspect warehouse - Local Warehouse connector manual

SYNOPSIS
  pm connectors inspect warehouse
  pm connectors inspect warehouse --json
  pm credentials add <name> --connector warehouse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Local JSONL warehouse destination used by the dependency-free MVP.

CAPABILITIES
  check=true catalog=true read=true write=true query=true
  Integration type: database

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  path: Local warehouse directory.

ETL STREAMS
  tables: Local JSONL warehouse tables.

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
  Source modes: full_refresh, incremental
  Destination modes: append, overwrite, append_dedup, overwrite_dedup

SECURITY
  read risk: local warehouse read
  write risk: local file write
  mutation risk: local dependency-free warehouse writes
  approval: not required for ETL destination writes; reverse ETL still requires approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect warehouse

  # Inspect as structured JSON
  pm connectors inspect warehouse --json

  # Warehouse credential
  pm credentials add warehouse-local --connector warehouse --config path=$ROOT/.polymetrics/warehouse
  pm query run --table sample_customers --limit 5 --json

AGENT WORKFLOW
  - Run pm connectors inspect warehouse before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
