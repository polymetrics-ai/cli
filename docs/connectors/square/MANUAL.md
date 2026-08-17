# pm connectors inspect square

```text
NAME
  pm connectors inspect square - Square connector manual

SYNOPSIS
  pm connectors inspect square
  pm connectors inspect square --json
  pm credentials add <name> --connector square [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Square payments, refunds, customers, and locations through the Square Connect v2 REST API.

ICON
  asset: icons/square.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.squareup.com/reference/square

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
  api_key (secret)

ETL STREAMS
  payments:
    primary key: id
    cursor: updated_at
    fields: amount_money(), created_at(), id(), location_id(), order_id(), processing_fee(), receipt_number(), source_type(), status(), total_money(), updated_at()
  refunds:
    primary key: id
    cursor: updated_at
    fields: amount_money(), created_at(), id(), location_id(), order_id(), payment_id(), processing_fee(), reason(), status(), updated_at()
  customers:
    primary key: id
    cursor: updated_at
    fields: company_name(), created_at(), creation_source(), email_address(), family_name(), given_name(), id(), phone_number(), reference_id(), updated_at()
  locations:
    primary key: id
    fields: country(), created_at(), currency(), id(), merchant_id(), name(), status(), timezone(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Square API read of payments, refunds, customer, and location data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect square

  # Inspect as structured JSON
  pm connectors inspect square --json

AGENT WORKFLOW
  - Run pm connectors inspect square before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
