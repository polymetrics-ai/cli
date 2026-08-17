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
  companyid
  mode
  page_size
  api_key (secret)

ETL STREAMS
  products:
    primary key: productId
    cursor: lastModifiedDateTime
    fields: categoryId(), description(), isActive(), isManufacturable(), itemType(), lastModifiedDateTime(), name(), productId(), sku(), timestamp(), trackSerials()
  customers:
    primary key: customerId
    fields: contactName(), customerId(), email(), fax(), isActive(), name(), phone(), pricingSchemeId(), remarks(), taxingSchemeId(), timestamp()
  vendors:
    primary key: vendorId
    fields: contactName(), currencyId(), email(), fax(), isActive(), leadTimeDays(), name(), phone(), taxingSchemeId(), timestamp(), vendorId()
  sales_orders:
    primary key: salesOrderId
    fields: amountPaid(), balance(), contactName(), currencyId(), customerId(), dueDate(), email(), inventoryStatus(), invoicedDate(), isCancelled(), isCompleted(), isInvoiced(), isQuote(), salesOrderId()
  categories:
    primary key: categoryId
    fields: categoryId(), isDefault(), name(), parentCategoryId(), timestamp()

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
