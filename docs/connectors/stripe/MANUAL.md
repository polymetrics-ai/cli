# pm connectors inspect stripe

```text
NAME
  pm connectors inspect stripe - Stripe connector manual

SYNOPSIS
  pm connectors inspect stripe
  pm connectors inspect stripe --json
  pm credentials add <name> --connector stripe [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse ETL customer actions through the Stripe REST API.

ICON
  asset: icons/stripe.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://stripe.com/docs/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  max_pages
  mode
  page_size
  start_date
  client_secret (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: created
    fields: balance(), created(), currency(), delinquent(), description(), email(), id(), livemode(), name(), object(), phone()
  charges:
    primary key: id
    cursor: created
    fields: amount(), amount_captured(), amount_refunded(), created(), currency(), customer(), id(), livemode(), object(), paid(), refunded(), status()
  invoices:
    primary key: id
    cursor: created
    fields: amount_due(), amount_paid(), amount_remaining(), created(), currency(), customer(), id(), livemode(), object(), paid(), status(), subscription(), total()
  subscriptions:
    primary key: id
    cursor: created
    fields: cancel_at_period_end(), canceled_at(), created(), currency(), current_period_end(), current_period_start(), customer(), id(), livemode(), object(), status()
  products:
    primary key: id
    cursor: created
    fields: active(), created(), description(), id(), livemode(), name(), object(), type(), updated()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: external mutation; approval required
  update_customer:
    endpoint: POST /customers/{{ record.id }}
    required fields: id
    risk: external mutation; approval required

SECURITY
  read risk: external Stripe API read of customer and billing data
  write risk: external Stripe API mutation
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect stripe

  # Inspect as structured JSON
  pm connectors inspect stripe --json

AGENT WORKFLOW
  - Run pm connectors inspect stripe before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
