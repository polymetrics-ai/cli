---
name: pm-freshsales
description: Freshsales connector knowledge and safe action guide.
---

# pm-freshsales

## Purpose

Reads Freshsales (Freshworks CRM) contacts, sales accounts, deals, and leads through the Freshsales REST API.

## Icon

- asset: icons/freshsales.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.freshworks.com/crm/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- domain_name
- max_pages
- mode
- view_id
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: city(), country(), created_at(), display_name(), email(), first_name(), id(), job_title(), last_name(), mobile_number(), owner_id(), updated_at(), work_number()
- sales_accounts:
  - primary key: id
  - cursor: updated_at
  - fields: annual_revenue(), city(), country(), created_at(), id(), industry_type_id(), name(), number_of_employees(), owner_id(), phone(), updated_at(), website()
- deals:
  - primary key: id
  - cursor: updated_at
  - fields: amount(), created_at(), currency_id(), deal_pipeline_id(), deal_stage_id(), expected_close(), id(), name(), owner_id(), probability(), sales_account_id(), updated_at()
- leads:
  - primary key: id
  - cursor: updated_at
  - fields: city(), company_name(), country(), created_at(), display_name(), email(), first_name(), id(), job_title(), last_name(), lead_stage_id(), owner_id(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Freshsales API read of CRM contact, account, deal, and lead data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshsales
```

### Inspect as structured JSON

```bash
pm connectors inspect freshsales --json
```

## Agent Rules

- Run pm connectors inspect freshsales before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
