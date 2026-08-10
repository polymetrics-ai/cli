# pm connectors inspect orb

```text
NAME
  pm connectors inspect orb - Orb connector manual

SYNOPSIS
  pm connectors inspect orb
  pm connectors inspect orb --json
  pm credentials add <name> --connector orb [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Orb customers, subscriptions, plans, and invoices.

ICON
  id: orb
  asset: icons/orb.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.withorb.com/reference/api-reference

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
  start_date
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: created_at
    fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
  subscriptions:
    primary key: id
    cursor: created_at
    fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
  plans:
    primary key: id
    cursor: created_at
    fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
  invoices:
    primary key: id
    cursor: created_at
    fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Orb API read of customer and billing data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect orb

  # Inspect as structured JSON
  pm connectors inspect orb --json

AGENT WORKFLOW
  - Run pm connectors inspect orb before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
