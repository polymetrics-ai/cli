# pm connectors inspect inflowinventory

```text
NAME
  pm connectors inspect inflowinventory - inFlow Inventory connector manual

SYNOPSIS
  pm connectors inspect inflowinventory
  pm connectors inspect inflowinventory --json
  pm credentials add <name> --connector inflowinventory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads inFlow Inventory products, customers, vendors, sales orders, and categories through the inFlow cloud REST API.

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
  companyid (required)
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  products:
    primary key: productId
    cursor: lastModifiedDateTime
    fields: categoryId(string), description(string), isActive(boolean), isManufacturable(boolean), itemType(string), lastModifiedDateTime(string), name(string), productId(string), sku(string), timestamp(string), trackSerials(boolean)
  customers:
    primary key: customerId
    fields: contactName(string), customerId(string), email(string), fax(string), isActive(boolean), name(string), phone(string), pricingSchemeId(string), remarks(string), taxingSchemeId(string), timestamp(string)
  vendors:
    primary key: vendorId
    fields: contactName(string), currencyId(string), email(string), fax(string), isActive(boolean), leadTimeDays(integer), name(string), phone(string), taxingSchemeId(string), timestamp(string), vendorId(string)
  sales_orders:
    primary key: salesOrderId
    fields: amountPaid(string), balance(string), contactName(string), currencyId(string), customerId(string), dueDate(string), email(string), inventoryStatus(string), invoicedDate(string), isCancelled(boolean), isCompleted(boolean), isInvoiced(boolean), isQuote(boolean), salesOrderId(string)
  categories:
    primary key: categoryId
    fields: categoryId(string), isDefault(boolean), name(string), parentCategoryId(string), timestamp(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external inFlow Inventory API read of products, customers, vendors, sales orders, and categories
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect inflowinventory

  # Inspect as structured JSON
  pm connectors inspect inflowinventory --json

AGENT WORKFLOW
  - Run pm connectors inspect inflowinventory before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
