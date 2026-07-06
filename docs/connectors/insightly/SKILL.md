---
name: pm-insightly
description: Insightly connector knowledge and safe action guide.
---

# pm-insightly

## Purpose

Reads Insightly CRM contacts, organisations, opportunities, leads, projects, and tasks through the Insightly REST API v3.1.

## Icon

- asset: icons/insightly.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- token (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: date_updated_utc
  - fields: contact_id(), date_created_utc(), date_updated_utc(), email_address(), first_name(), id(), last_name(), organisation_id(), phone(), title()
- organisations:
  - primary key: id
  - cursor: date_updated_utc
  - fields: date_created_utc(), date_updated_utc(), id(), organisation_id(), organisation_name(), owner_user_id(), phone(), website()
- opportunities:
  - primary key: id
  - cursor: date_updated_utc
  - fields: bid_currency(), date_created_utc(), date_updated_utc(), id(), opportunity_id(), opportunity_name(), opportunity_state(), opportunity_value(), pipeline_id(), probability(), stage_id()
- leads:
  - primary key: id
  - cursor: date_updated_utc
  - fields: converted(), date_created_utc(), date_updated_utc(), email(), first_name(), id(), last_name(), lead_id(), lead_source_id(), lead_status_id(), organisation_name()
- projects:
  - primary key: id
  - cursor: date_updated_utc
  - fields: date_created_utc(), date_updated_utc(), id(), owner_user_id(), pipeline_id(), project_id(), project_name(), stage_id(), status()
- tasks:
  - primary key: id
  - cursor: date_updated_utc
  - fields: completed(), date_created_utc(), date_updated_utc(), due_date(), id(), owner_user_id(), priority(), status(), task_id(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Insightly API read of contacts, organisations, opportunities, leads, projects, and tasks
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect insightly
```

### Inspect as structured JSON

```bash
pm connectors inspect insightly --json
```

## Agent Rules

- Run pm connectors inspect insightly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
