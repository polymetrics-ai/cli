---
name: pm-klarna
description: Klarna connector knowledge and safe action guide.
---

# pm-klarna

## Purpose

Reads Klarna settlement payouts and transactions through the Klarna Settlements API.

## Icon

- asset: icons/klarna.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.klarna.com/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- payment_references
- summary_currency_code
- summary_end_date
- summary_start_date
- password (secret)
- username (secret)

## ETL Streams

- payouts:
  - primary key: payout_reference
  - fields: currency_code(), merchant_settlement_type(), payment_reference(), payout_reference(), settlement_amount(), totals()
- transactions:
  - primary key: transaction_id
  - fields: amount(), capture_date(), capture_id(), currency_code(), merchant_reference1(), merchant_reference2(), order_id(), payout_reference(), sale_date(), short_order_id(), transaction_id(), type()
- payout_details:
  - primary key: payment_reference
  - fields: currency_code(), currency_code_of_registration_country(), merchant_id(), merchant_settlement_type(), payment_reference(), payout_date(), totals(), transactions()
- payout_summaries:
  - primary key: summary_settlement_currency, summary_payout_date_start, summary_payout_date_end
  - fields: summary_payout_date_end(), summary_payout_date_start(), summary_settlement_currency(), summary_total_commission_amount(), summary_total_commission_reversal_amount(), summary_total_fee_amount(), summary_total_fee_correction_amount(), summary_total_holdback_amount(), summary_total_release_amount(), summary_total_repay_amount(), summary_total_return_amount(), summary_total_reversal_amount(), summary_total_sale_amount(), summary_total_settlement_amount(), summary_total_tax_amount()
- payout_summary:
  - primary key: payout_reference
  - fields: currency_code(), fee_amount(), payout_reference(), return_amount(), sale_amount(), settlement_amount()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Klarna Settlements API read of payout and transaction data
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect klarna
```

### Inspect as structured JSON

```bash
pm connectors inspect klarna --json
```

## Agent Rules

- Run pm connectors inspect klarna before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
