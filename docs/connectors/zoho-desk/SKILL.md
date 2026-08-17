---
name: pm-zoho-desk
description: Zoho Desk connector knowledge and safe action guide.
---

# pm-zoho-desk

## Purpose

Reads Zoho Desk tickets, contacts, and accounts through the Zoho Desk REST API.

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
- org_id
- page_size
- access_token (secret)

## ETL Streams

- tickets:
  - primary key: id
  - cursor: updated_at
  - fields: channel(), createdTime(), email(), id(), modifiedTime(), name(), priority(), status(), subject(), ticketNumber(), updated_at()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: accountId(), createdTime(), email(), firstName(), id(), lastName(), modifiedTime(), name(), phone(), updated_at()
- accounts:
  - primary key: id
  - cursor: updated_at
  - fields: accountName(), createdTime(), id(), modifiedTime(), name(), phone(), updated_at(), website()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoho Desk API read of support ticket and contact data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-desk
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-desk --json
```

## Agent Rules

- Run pm connectors inspect zoho-desk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
