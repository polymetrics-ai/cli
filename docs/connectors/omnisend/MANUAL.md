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
  api_key (secret) (required)

ETL STREAMS
  contacts:
    primary key: contactID
    cursor: createdAt
    fields: city(string), contactID(string), country(string), countryCode(string), createdAt(string), email(string), firstName(string), lastName(string), segments(array), state(string), status(string), tags(array)
  campaigns:
    primary key: campaignID
    cursor: createdAt
    fields: bounced(number), campaignID(string), clicked(number), createdAt(string), endDate(string), fromName(string), name(string), opened(number), sent(number), startDate(string), status(string), subject(string), type(string), unsubscribed(number)
  carts:
    primary key: cartID
    cursor: createdAt
    fields: cartID(string), cartRecoveryUrl(string), cartSum(number), contactID(string), createdAt(string), currency(string), email(string), phone(string), products(array), updatedAt(string)
  orders:
    primary key: orderID
    cursor: createdAt
    fields: cartID(string), contactID(string), createdAt(string), currency(string), discountSum(number), email(string), fulfillmentStatus(string), orderID(string), orderNumber(number), orderSum(number), paymentStatus(string), products(array), shippingSum(number), subTotalSum(number), taxSum(number), updatedAt(string)
  products:
    primary key: productID
    cursor: createdAt
    fields: categoryIDs(array), createdAt(string), currency(string), description(string), images(array), productID(string), productUrl(string), status(string), title(string), type(string), updatedAt(string), variants(array), vendor(string)

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
