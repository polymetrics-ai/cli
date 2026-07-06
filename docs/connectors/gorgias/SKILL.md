---
name: pm-gorgias
description: Gorgias connector knowledge and safe action guide.
---

# pm-gorgias

## Purpose

Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias REST API (read-only).

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
- page_size
- username
- password (secret)

## ETL Streams

- tickets:
  - primary key: id
  - cursor: updated_datetime
  - fields: channel(), closed_datetime(), created_datetime(), id(), is_unread(), language(), opened_datetime(), priority(), spam(), status(), subject(), trashed_datetime(), updated_datetime(), via()
- customers:
  - primary key: id
  - cursor: updated_datetime
  - fields: channel(), created_datetime(), email(), external_id(), firstname(), id(), language(), lastname(), name(), timezone(), updated_datetime()
- messages:
  - primary key: id
  - cursor: created_datetime
  - fields: body_text(), channel(), created_datetime(), from_agent(), id(), public(), sent_datetime(), stripped_text(), subject(), ticket_id(), via()
- satisfaction_surveys:
  - primary key: id
  - cursor: created_datetime
  - fields: body_text(), created_datetime(), customer_id(), id(), scale_range(), score(), scored_datetime(), sent_datetime(), ticket_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Gorgias API read of helpdesk tickets, customers, messages, and satisfaction surveys
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gorgias
```

### Inspect as structured JSON

```bash
pm connectors inspect gorgias --json
```

## Agent Rules

- Run pm connectors inspect gorgias before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
