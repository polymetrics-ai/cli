# pm connectors inspect lob

```text
NAME
  pm connectors inspect lob - Lob connector manual

SYNOPSIS
  pm connectors inspect lob
  pm connectors inspect lob --json
  pm credentials add <name> --connector lob [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail REST API.

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
  api_key (secret)

ETL STREAMS
  addresses:
    primary key: id
    cursor: date_created
    fields: address_city(), address_country(), address_line1(), address_line2(), address_state(), address_zip(), company(), date_created(), date_modified(), deleted(), description(), email(), id(), name(), object(), phone()
  postcards:
    primary key: id
    cursor: date_created
    fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
  letters:
    primary key: id
    cursor: date_created
    fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
  checks:
    primary key: id
    cursor: date_created
    fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
  bank_accounts:
    primary key: id
    cursor: date_created
    fields: account_number(), account_type(), bank_name(), date_created(), date_modified(), deleted(), description(), id(), object(), routing_number(), signatory(), verified()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Lob API read of address book, mailpiece, and bank account data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect lob

  # Inspect as structured JSON
  pm connectors inspect lob --json

AGENT WORKFLOW
  - Run pm connectors inspect lob before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
