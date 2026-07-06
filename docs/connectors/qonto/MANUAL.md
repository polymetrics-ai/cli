# pm connectors inspect qonto

```text
NAME
  pm connectors inspect qonto - Qonto connector manual

SYNOPSIS
  pm connectors inspect qonto
  pm connectors inspect qonto --json
  pm credentials add <name> --connector qonto [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Qonto bank transactions, memberships, and accounts through the Qonto REST API (read-only).

ICON
  asset: icons/qonto.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api-doc.qonto.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  iban
  start_date
  api_key (secret)

ETL STREAMS
  transactions:
    primary key: id
    cursor: settled_at
    fields: amount(), id(), settled_at(), side(), updated_at()
  memberships:
    primary key: id
    fields: amount(), id(), settled_at(), side(), updated_at()
  accounts:
    primary key: id
    fields: amount(), id(), settled_at(), side(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Qonto API read of bank transaction and account data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect qonto

  # Inspect as structured JSON
  pm connectors inspect qonto --json

AGENT WORKFLOW
  - Run pm connectors inspect qonto before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
