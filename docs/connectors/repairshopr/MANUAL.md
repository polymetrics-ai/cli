# pm connectors inspect repairshopr

```text
NAME
  pm connectors inspect repairshopr - RepairShopr connector manual

SYNOPSIS
  pm connectors inspect repairshopr
  pm connectors inspect repairshopr --json
  pm credentials add <name> --connector repairshopr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads RepairShopr customers, tickets, invoices, estimates, and assets through the REST API.

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
  created_after
  query
  updated_after
  api_token (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), name(), phone(), stream(), updated_at()
  tickets:
    primary key: id
    cursor: updated_at
    fields: customer_id(), id(), number(), status(), stream(), subject(), updated_at()
  invoices:
    primary key: id
    cursor: updated_at
    fields: customer_id(), id(), number(), status(), stream(), total(), updated_at()
  estimates:
    primary key: id
    cursor: updated_at
    fields: customer_id(), id(), number(), status(), stream(), total(), updated_at()
  assets:
    primary key: id
    cursor: updated_at
    fields: customer_id(), id(), name(), serial_number(), stream(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external RepairShopr API read of customer and shop-management data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect repairshopr

  # Inspect as structured JSON
  pm connectors inspect repairshopr --json

AGENT WORKFLOW
  - Run pm connectors inspect repairshopr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
