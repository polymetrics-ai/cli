# pm connectors inspect squarespace

```text
NAME
  pm connectors inspect squarespace - Squarespace connector manual

SYNOPSIS
  pm connectors inspect squarespace
  pm connectors inspect squarespace --json
  pm credentials add <name> --connector squarespace [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace Commerce API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(), id(), modifiedOn(), orderNumber()
  products:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(), id(), modifiedOn(), name()
  inventory:
    primary key: sku
    fields: modifiedOn(), quantity(), sku()
  profiles:
    primary key: id
    fields: createdOn(), id(), modifiedOn(), name()
  transactions:
    primary key: id
    fields: createdOn(), customerEmail(), discounts(), id(), modifiedOn(), payments(), salesLineItems(), salesOrderId(), shippingLineItems(), total(), totalNetPayment(), totalNetSales(), totalNetShipping(), totalSales(), totalTaxes(), voided()
  store_pages:
    primary key: id
    fields: id(), isEnabled(), title(), urlSlug()
  webhook_subscriptions:
    primary key: id
    fields: clientId(), createdOn(), endpointUrl(), id(), topics(), updatedOn(), websiteId()
  contacts:
    primary key: id
    fields: createdOn(), defaultShippingAddress(), firstName(), id(), lastName(), locale(), primaryEmail()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_webhook_subscription:
    endpoint: POST /webhook_subscriptions
    risk: registers a new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external mutation, no approval required
  delete_webhook_subscription:
    endpoint: DELETE /webhook_subscriptions/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription, stopping future event notifications to that endpoint; external mutation, approval required

SECURITY
  read risk: external Squarespace API read of commerce orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts
  write risk: external Squarespace API mutation (webhook subscription create/delete)
  approval: reverse ETL plan approval required before destructive writes (delete_webhook_subscription); create_webhook_subscription is low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect squarespace

  # Inspect as structured JSON
  pm connectors inspect squarespace --json

AGENT WORKFLOW
  - Run pm connectors inspect squarespace before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
