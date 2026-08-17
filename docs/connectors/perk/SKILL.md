---
name: pm-perk
description: Perk connector knowledge and safe action guide.
---

# pm-perk

## Purpose

Reads Perk/TravelPerk trips and invoices through read-only REST list endpoints.

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
- page_size
- start_date
- api_key (secret)

## ETL Streams

- trips:
  - primary key: id
  - cursor: modified
  - fields: id(), modified(), status(), trip_name()
- invoices:
  - primary key: serial_number
  - cursor: issuing_date
  - fields: issuing_date(), serial_number(), status(), total()
- invoice_lines:
  - primary key: id
  - cursor: issuing_date
  - fields: currency(), description(), due_date(), expense_date(), id(), invoice_mode(), invoice_serial_number(), invoice_status(), issuing_date(), metadata(), profile_id(), profile_name(), quantity(), tax_amount(), tax_percentage(), tax_regime(), total_amount(), unit_price()
- invoice_profiles:
  - primary key: id
  - fields: billing_information(), billing_period(), currency(), id(), name(), payment_method_type()
- trip_custom_fields:
  - primary key: trip_id
  - fields: created_date(), custom_fields(), trip_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Perk/TravelPerk API read of trip and invoice data
- approval: none; read-only, no writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect perk
```

### Inspect as structured JSON

```bash
pm connectors inspect perk --json
```

## Agent Rules

- Run pm connectors inspect perk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
