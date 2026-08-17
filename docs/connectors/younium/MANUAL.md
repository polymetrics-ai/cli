# pm connectors inspect younium

```text
NAME
  pm connectors inspect younium - Younium connector manual

SYNOPSIS
  pm connectors inspect younium
  pm connectors inspect younium --json
  pm credentials add <name> --connector younium [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Younium accounts, subscriptions, invoices, products, payment terms, currencies, and webhooks through the Younium REST API.

ICON
  asset: icons/younium.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.younium.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  legal_entity
  mode
  username
  password (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: updated_at
    fields: account_id(), id(), name(), updated_at()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: account_id(), id(), name(), updated_at()
  invoices:
    primary key: id
    cursor: updated_at
    fields: account_id(), id(), name(), updated_at()
  products:
    primary key: id
    fields: id(), name(), updated_at()
  payment_terms:
    primary key: id
    fields: id(), name()
  currencies:
    primary key: id
    fields: id(), name()
  webhooks:
    primary key: id
    fields: id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /Accounts
    risk: creates a new billing account in Younium; external mutation, approval required
  update_account:
    endpoint: PATCH /Accounts/{{ record.id }}
    required fields: id
    risk: mutates an existing account's billing/contact/tax metadata; external mutation, approval required
  cancel_subscription:
    endpoint: POST /Subscriptions/cancel/{{ record.id }}
    required fields: id
    risk: irreversibly schedules or immediately cancels an active subscription, ending future billing; external mutation, approval required
  post_invoice:
    endpoint: POST /Invoices/{{ record.id }}/post
    required fields: id
    risk: finalizes a draft invoice, making it official/sendable to the customer; external mutation, approval required
  cancel_invoice:
    endpoint: POST /Invoices/{{ record.id }}/cancel
    required fields: id
    risk: irreversibly cancels a posted invoice; external mutation, approval required

SECURITY
  read risk: external Younium API read of account, subscription, invoice, product, payment term, currency, and webhook data
  write risk: external mutation of billing-critical Younium records: account create/update, subscription cancellation (ends future billing), and invoice posting/cancellation
  approval: required for all write actions; cancel_subscription and cancel_invoice require explicit destructive confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect younium

  # Inspect as structured JSON
  pm connectors inspect younium --json

AGENT WORKFLOW
  - Run pm connectors inspect younium before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
