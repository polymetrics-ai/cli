---
name: pm-ebay-finance
description: eBay Finance connector knowledge and safe action guide.
---

# pm-ebay-finance

## Purpose

Reads eBay seller financial data — transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- client_access_token (secret)

## ETL Streams

- transactions:
  - primary key: transactionId
  - cursor: transactionDate
  - fields: amount_currency(), amount_value(), bookingEntry(), feeType(), orderId(), payoutId(), salesRecordReference(), transactionDate(), transactionId(), transactionMemo(), transactionStatus(), transactionType()
- payouts:
  - primary key: payoutId
  - cursor: payoutDate
  - fields: amount_currency(), amount_value(), payoutDate(), payoutId(), payoutInstrument_accountLastFourDigits(), payoutInstrument_nickname(), payoutStatus(), payoutStatusDescription(), transactionCount()
- transfers:
  - primary key: transferId
  - cursor: transferDate
  - fields: amount_currency(), amount_value(), reason(), transferDate(), transferId(), transferStatus(), transferType()
- seller_funds_summary:
  - fields: availableFunds_currency(), availableFunds_value(), fundsOnHold_currency(), fundsOnHold_value(), processingFunds_currency(), processingFunds_value(), totalFunds_currency(), totalFunds_value()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external eBay Sell Finances API read of a seller's monetary records
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ebay-finance
```

### Inspect as structured JSON

```bash
pm connectors inspect ebay-finance --json
```

## Agent Rules

- Run pm connectors inspect ebay-finance before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
