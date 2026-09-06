# pm connectors inspect fastbill

```text
NAME
  pm connectors inspect fastbill - FastBill connector manual

SYNOPSIS
  pm connectors inspect fastbill
  pm connectors inspect fastbill --json
  pm credentials add <name> --connector fastbill [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FastBill billing records through fixed JSON SERVICE envelopes.

ICON
  id: fastbill
  asset: icons/fastbill.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.fastbill.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  username (required)
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: CUSTOMER_ID
    fields: CUSTOMER_ID(string)
  invoices:
    primary key: INVOICE_ID
    fields: INVOICE_ID(string)
  products:
    primary key: ARTICLE_NUMBER
    fields: ARTICLE_NUMBER(string)
  recurring_invoices:
    primary key: INVOICE_ID
    fields: INVOICE_ID(string)
  revenues:
    primary key: REVENUE_ID
    fields: REVENUE_ID(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: Bounded FastBill JSON API requests use fixed origin and declared Basic authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fastbill

  # Inspect as structured JSON
  pm connectors inspect fastbill --json

AGENT WORKFLOW
  - Run pm connectors inspect fastbill before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
