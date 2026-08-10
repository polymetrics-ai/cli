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
  id: simple-icons-squarespace
  asset: icons/simple-icons/squarespace.svg
  title: Squarespace
  simple_icon_slug: squarespace
  simple_icon_hex: 000000
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Squarespace
  match: exact-name-or-slug
  matched_by: squarespace

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  orders:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(string), id(string), modifiedOn(string), orderNumber(string)
  products:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(string), id(string), modifiedOn(string), name(string)
  inventory:
    primary key: sku
    fields: modifiedOn(string), quantity(integer), sku(string)
  profiles:
    primary key: id
    fields: createdOn(string), id(string), modifiedOn(string), name(string)
  transactions:
    primary key: id
    fields: createdOn(string), customerEmail(string), discounts(array), id(string), modifiedOn(string), payments(array), salesLineItems(array), salesOrderId(string), shippingLineItems(array), total(object), totalNetPayment(object), totalNetSales(object), totalNetShipping(object), totalSales(object), totalTaxes(object), voided(boolean)
  store_pages:
    primary key: id
    fields: id(string), isEnabled(boolean), title(string), urlSlug(string)
  webhook_subscriptions:
    primary key: id
    fields: clientId(string), createdOn(string), endpointUrl(string), id(string), topics(array), updatedOn(string), websiteId(string)
  contacts:
    primary key: id
    fields: createdOn(string), defaultShippingAddress(object), firstName(string), id(string), lastName(string), locale(string), primaryEmail(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_webhook_subscription:
    endpoint: POST /webhook_subscriptions
    required fields: endpointUrl
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
