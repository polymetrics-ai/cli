---
name: pm-zoho-invoice
description: Zoho Invoice connector knowledge and safe action guide.
---

# pm-zoho-invoice

## Purpose

Reads Zoho Invoice customers, invoices, and payments through the Zoho Invoice REST API.

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
- max_pages
- mode
- organization_id
- page_size
- access_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: company_name(), created_time(), currency_code(), customer_id(), customer_name(), customer_type(), email(), id(), last_modified_time(), outstanding_receivable_amount(), phone(), status(), updated_at()
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: balance(), created_time(), currency_code(), customer_id(), customer_name(), date(), due_date(), id(), invoice_id(), invoice_number(), last_modified_time(), status(), total(), updated_at()
- payments:
  - primary key: id
  - cursor: updated_at
  - fields: amount(), created_time(), currency_code(), customer_id(), customer_name(), date(), id(), invoice_numbers(), last_modified_time(), payment_id(), payment_mode(), payment_number(), reference_number(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoho Invoice API read of customer/invoice/payment data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-invoice
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-invoice --json
```

## Agent Rules

- Run pm connectors inspect zoho-invoice before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
