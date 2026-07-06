---
name: pm-freshdesk
description: Freshdesk connector knowledge and safe action guide.
---

# pm-freshdesk

## Purpose

Reads Freshdesk tickets, contacts, companies, agents, and groups through the Freshdesk REST API v2.

## Icon

- asset: icons/freshdesk.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.freshdesk.com/api/#change_log

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

- tickets:
  - primary key: id
  - cursor: updated_at
  - fields: company_id(), created_at(), due_by(), group_id(), id(), priority(), requester_id(), responder_id(), source(), spam(), status(), subject(), type(), updated_at()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: active(), company_id(), created_at(), email(), id(), mobile(), name(), phone(), updated_at()
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), note(), updated_at()
- agents:
  - primary key: id
  - cursor: updated_at
  - fields: available(), created_at(), id(), occasional(), ticket_scope(), updated_at()
- groups:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Freshdesk API read of support tickets, contacts, companies, agents, and groups
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshdesk
```

### Inspect as structured JSON

```bash
pm connectors inspect freshdesk --json
```

## Agent Rules

- Run pm connectors inspect freshdesk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
