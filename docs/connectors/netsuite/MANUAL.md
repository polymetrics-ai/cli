# pm connectors inspect netsuite

```text
NAME
  pm connectors inspect netsuite - NetSuite connector manual

SYNOPSIS
  pm connectors inspect netsuite
  pm connectors inspect netsuite --json
  pm credentials add <name> --connector netsuite [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads selected NetSuite REST Record API resources (customers, vendors, items, sales orders), authenticating with OAuth 1.0a Token-Based Authentication (HMAC-SHA256 request signing). Read-only.

ICON
  asset: icons/netsuite.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html

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
  realm
  consumer_key (secret)
  consumer_secret (secret)
  token_key (secret)
  token_secret (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: last_modified_date
    fields: email(), entity_id(), id(), last_modified_date(), name(), status()
  vendors:
    primary key: id
    cursor: last_modified_date
    fields: email(), entity_id(), id(), last_modified_date(), name(), status()
  items:
    primary key: id
    cursor: last_modified_date
    fields: email(), entity_id(), id(), last_modified_date(), name(), status()
  sales_orders:
    primary key: id
    cursor: last_modified_date
    fields: email(), entity_id(), id(), last_modified_date(), name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NetSuite REST Record API read of customer, vendor, item, and sales order data
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect netsuite

  # Inspect as structured JSON
  pm connectors inspect netsuite --json

AGENT WORKFLOW
  - Run pm connectors inspect netsuite before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
