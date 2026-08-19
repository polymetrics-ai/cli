---
name: pm-ebay-finance
description: eBay Finance connector knowledge and safe action guide.
---

# pm-ebay-finance

## Purpose

Reads eBay seller financial data — transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances REST API.

## Icon

- id: simple-icons-ebay-finance
- asset: icons/simple-icons/ebay-finance.svg
- title: eBay
- simple_icon_slug: ebay
- simple_icon_hex: E53238
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=eBay
- match: curated-alias
- matched_by: ebay

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
  - fields: amount_currency(string), amount_value(string), bookingEntry(string), feeType(string), orderId(string), payoutId(string), salesRecordReference(string), transactionDate(string), transactionId(string), transactionMemo(string), transactionStatus(string), transactionType(string)
- payouts:
  - primary key: payoutId
  - cursor: payoutDate
  - fields: amount_currency(string), amount_value(string), payoutDate(string), payoutId(string), payoutInstrument_accountLastFourDigits(string), payoutInstrument_nickname(string), payoutStatus(string), payoutStatusDescription(string), transactionCount(integer)
- transfers:
  - primary key: transferId
  - cursor: transferDate
  - fields: amount_currency(string), amount_value(string), reason(string), transferDate(string), transferId(string), transferStatus(string), transferType(string)
- seller_funds_summary:
  - fields: availableFunds_currency(string), availableFunds_value(string), fundsOnHold_currency(string), fundsOnHold_value(string), processingFunds_currency(string), processingFunds_value(string), totalFunds_currency(string), totalFunds_value(string)

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
