---
name: pm-tyntec-sms
description: tyntec SMS connector knowledge and safe action guide.
---

# pm-tyntec-sms

## Purpose

Reads tyntec SMS messages, templates, sender IDs, and delivery reports through API list endpoints, and sends approved SMS messages through the Messaging API.

## Icon

- asset: icons/tyntec.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.tyntec.com/reference/messaging

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- messages:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), from(), id(), status(), to()
- templates:
  - primary key: id
  - fields: id(), name()
- sender_ids:
  - primary key: id
  - fields: id(), name()
- delivery_reports:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), from(), id(), status(), to()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- send_message:
  - endpoint: POST sms/v1/messages
  - risk: sends a billable SMS message to the recipient phone number and may notify an external user

## Security

- read risk: external tyntec SMS API read of messages, templates, sender IDs, and delivery reports
- write risk: sends billable SMS messages to recipient phone numbers; approval required before delivery
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tyntec-sms
```

### Inspect as structured JSON

```bash
pm connectors inspect tyntec-sms --json
```

## Agent Rules

- Run pm connectors inspect tyntec-sms before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
