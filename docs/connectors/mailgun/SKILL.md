---
name: pm-mailgun
description: Mailgun connector knowledge and safe action guide.
---

# pm-mailgun

## Purpose

Reads Mailgun sending domains, email events, mailing lists, and analytics tags through the Mailgun v3 REST API.

## Icon

- asset: icons/mailgun.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://documentation.mailgun.com/en/latest/api_reference.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- domain_name
- mode
- page_size
- private_key (secret)

## ETL Streams

- domains:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), is_disabled(), name(), smtp_login(), spam_action(), state(), type(), wildcard()
- events:
  - primary key: id
  - cursor: timestamp
  - fields: event(), id(), log_level(), message_id(), reason(), recipient(), timestamp()
- mailing_lists:
  - primary key: address
  - cursor: created_at
  - fields: access_level(), address(), created_at(), description(), members_count(), name()
- tags:
  - primary key: tag
  - fields: description(), first_seen(), last_seen(), tag()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mailgun API read of sending-domain, event, mailing-list, and tag data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailgun
```

### Inspect as structured JSON

```bash
pm connectors inspect mailgun --json
```

## Agent Rules

- Run pm connectors inspect mailgun before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
