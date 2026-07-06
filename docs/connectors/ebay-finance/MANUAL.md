# pm connectors inspect ebay-finance

```text
NAME
  pm connectors inspect ebay-finance - eBay Finance connector manual

SYNOPSIS
  pm connectors inspect ebay-finance
  pm connectors inspect ebay-finance --json
  pm credentials add <name> --connector ebay-finance [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads eBay seller financial data — transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances REST API.

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
  mode
  start_date
  client_access_token (secret)

ETL STREAMS
  transactions:
    primary key: transactionId
    cursor: transactionDate
    fields: amount_currency(), amount_value(), bookingEntry(), feeType(), orderId(), payoutId(), salesRecordReference(), transactionDate(), transactionId(), transactionMemo(), transactionStatus(), transactionType()
  payouts:
    primary key: payoutId
    cursor: payoutDate
    fields: amount_currency(), amount_value(), payoutDate(), payoutId(), payoutInstrument_accountLastFourDigits(), payoutInstrument_nickname(), payoutStatus(), payoutStatusDescription(), transactionCount()
  transfers:
    primary key: transferId
    cursor: transferDate
    fields: amount_currency(), amount_value(), reason(), transferDate(), transferId(), transferStatus(), transferType()
  seller_funds_summary:
    fields: availableFunds_currency(), availableFunds_value(), fundsOnHold_currency(), fundsOnHold_value(), processingFunds_currency(), processingFunds_value(), totalFunds_currency(), totalFunds_value()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external eBay Sell Finances API read of a seller's monetary records
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ebay-finance

  # Inspect as structured JSON
  pm connectors inspect ebay-finance --json

AGENT WORKFLOW
  - Run pm connectors inspect ebay-finance before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
