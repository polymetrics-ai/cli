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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

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
  trips:
    primary key: id
    cursor: modified
    fields: id(string), modified(string), status(string), trip_name(string)
  invoices:
    primary key: serial_number
    cursor: issuing_date
    fields: issuing_date(string), serial_number(string), status(string), total(string)
  invoice_lines:
    primary key: id
    cursor: issuing_date
    fields: currency(string), description(string), due_date(string), expense_date(string), id(string), invoice_mode(string), invoice_serial_number(string), invoice_status(string), issuing_date(string), metadata(object), profile_id(string), profile_name(string), quantity(integer), tax_amount(string), tax_percentage(string), tax_regime(string), total_amount(string), unit_price(string)
  invoice_profiles:
    primary key: id
    fields: billing_information(object), billing_period(string), currency(string), id(string), name(string), payment_method_type(string)
  trip_custom_fields:
    primary key: trip_id
    fields: created_date(string), custom_fields(array), trip_id(string)

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
