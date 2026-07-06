# pm connectors inspect perk

```text
NAME
  pm connectors inspect perk - Perk connector manual

SYNOPSIS
  pm connectors inspect perk
  pm connectors inspect perk --json
  pm credentials add <name> --connector perk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Perk/TravelPerk trips and invoices through read-only REST list endpoints.

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
  max_pages
  mode
  page_size
  start_date
  api_key (secret)

ETL STREAMS
  trips:
    primary key: id
    cursor: modified
    fields: id(), modified(), status(), trip_name()
  invoices:
    primary key: serial_number
    cursor: issuing_date
    fields: issuing_date(), serial_number(), status(), total()
  invoice_lines:
    primary key: id
    cursor: issuing_date
    fields: currency(), description(), due_date(), expense_date(), id(), invoice_mode(), invoice_serial_number(), invoice_status(), issuing_date(), metadata(), profile_id(), profile_name(), quantity(), tax_amount(), tax_percentage(), tax_regime(), total_amount(), unit_price()
  invoice_profiles:
    primary key: id
    fields: billing_information(), billing_period(), currency(), id(), name(), payment_method_type()
  trip_custom_fields:
    primary key: trip_id
    fields: created_date(), custom_fields(), trip_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Perk/TravelPerk API read of trip and invoice data
  approval: none; read-only, no writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect perk

  # Inspect as structured JSON
  pm connectors inspect perk --json

AGENT WORKFLOW
  - Run pm connectors inspect perk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
