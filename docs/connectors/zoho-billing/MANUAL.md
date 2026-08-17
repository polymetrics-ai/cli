# pm connectors inspect zoho-billing

```text
NAME
  pm connectors inspect zoho-billing - Zoho Billing connector manual

SYNOPSIS
  pm connectors inspect zoho-billing
  pm connectors inspect zoho-billing --json
  pm credentials add <name> --connector zoho-billing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoho Billing customers, subscriptions, and invoices through the Zoho Billing REST API.

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
  organization_id
  access_token (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: customer_id(), display_name(), id(), name(), status(), updated_at(), updated_time()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), subscription_id(), updated_at(), updated_time()
  invoices:
    primary key: id
    cursor: updated_at
    fields: id(), invoice_id(), invoice_number(), name(), status(), updated_at(), updated_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoho Billing API read of customer and billing data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-billing

  # Inspect as structured JSON
  pm connectors inspect zoho-billing --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-billing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
