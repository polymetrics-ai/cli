---
name: pm-xsolla
description: Xsolla connector knowledge and safe action guide.
---

# pm-xsolla

## Purpose

Reads Xsolla merchant transaction search/registry, payouts, payout currency breakdown, and financial report data, and writes full/partial transaction refunds through the Xsolla Pay Station API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- datetime_from
- datetime_to
- merchant_id
- mode
- project_id
- api_key (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), updated_at()
- orders:
  - primary key: id
  - cursor: updated_at
  - fields: id(), status(), updated_at()
- transactions:
  - primary key: id
  - cursor: updated_at
  - fields: id(), status(), updated_at()
- transactions_search:
  - primary key: transaction_id
  - cursor: transaction_create_date
  - fields: payment_details(), payment_system(), purchase(), transaction(), transaction_create_date(), transaction_id(), user()
- transactions_registry:
  - primary key: transaction_id
  - cursor: transaction_transfer_date
  - fields: purchase(), transaction(), transaction_id(), transaction_transfer_date(), user(), user_balance()
- payouts:
  - primary key: payout_id
  - cursor: payout_date
  - fields: canceled(), payout(), payout_date(), payout_id(), rate(), transfer()
- payout_currency_breakdown:
  - primary key: IsoCurrency
  - fields: DirectTaxesOfPayments(), IsoCurrency(), PaymentsAmount(), SumCommissionAgent(), SumCommissionUserTaxes(), SumItems(), SumNominalSum(), SumOutProject(), SumPayoutSum(), TaxesOfPayments()
- financial_reports:
  - primary key: report_id
  - fields: agreement_document_id(), currency(), is_direct_payout(), is_draft_by_agreement(), month(), report_id(), year()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- request_refund:
  - endpoint: PUT /merchants/{{ config.merchant_id }}/reports/transactions/{{ record.transaction_id }}/refund
  - required fields: transaction_id
  - risk: irreversible external mutation; issues a full refund to the user for the given transaction; approval required
- request_partial_refund:
  - endpoint: PUT /merchants/{{ config.merchant_id }}/reports/transactions/{{ record.transaction_id }}/partial_refund
  - required fields: transaction_id
  - risk: irreversible external mutation; issues a partial refund to the user for the given transaction; approval required

## Security

- read risk: external Xsolla merchant API read of transaction, payout, and financial report data
- write risk: external mutation: issues full or partial refunds to end users for completed transactions
- approval: required for all write actions (request_refund, request_partial_refund); reads require none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect xsolla
```

### Inspect as structured JSON

```bash
pm connectors inspect xsolla --json
```

## Agent Rules

- Run pm connectors inspect xsolla before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
