# pm connectors inspect omnisend

```text
NAME
  pm connectors inspect omnisend - Omnisend connector manual

SYNOPSIS
  pm connectors inspect omnisend
  pm connectors inspect omnisend --json
  pm credentials add <name> --connector omnisend [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Omnisend contacts, campaigns, carts, orders, and products through the Omnisend REST API.

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
  contacts:
    primary key: contactID
    cursor: createdAt
    fields: city(), contactID(), country(), countryCode(), createdAt(), email(), firstName(), lastName(), segments(), state(), status(), tags()
  campaigns:
    primary key: campaignID
    cursor: createdAt
    fields: bounced(), campaignID(), clicked(), createdAt(), endDate(), fromName(), name(), opened(), sent(), startDate(), status(), subject(), type(), unsubscribed()
  carts:
    primary key: cartID
    cursor: createdAt
    fields: cartID(), cartRecoveryUrl(), cartSum(), contactID(), createdAt(), currency(), email(), phone(), products(), updatedAt()
  orders:
    primary key: orderID
    cursor: createdAt
    fields: cartID(), contactID(), createdAt(), currency(), discountSum(), email(), fulfillmentStatus(), orderID(), orderNumber(), orderSum(), paymentStatus(), products(), shippingSum(), subTotalSum(), taxSum(), updatedAt()
  products:
    primary key: productID
    cursor: createdAt
    fields: categoryIDs(), createdAt(), currency(), description(), images(), productID(), productUrl(), status(), title(), type(), updatedAt(), variants(), vendor()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Omnisend API read of contact, campaign, and ecommerce order data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect omnisend

  # Inspect as structured JSON
  pm connectors inspect omnisend --json

AGENT WORKFLOW
  - Run pm connectors inspect omnisend before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
