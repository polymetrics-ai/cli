---
name: pm-sendgrid
description: Sendgrid connector knowledge and safe action guide.
---

# pm-sendgrid

## Purpose

Reads SendGrid Marketing Campaigns lists, segments, and contacts, plus suppression bounces, through the SendGrid v3 REST API. Read-only.

## Icon

- asset: icons/sendgrid.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- lists:
  - primary key: id
  - fields: contact_count(), id(), name()
- segments:
  - primary key: id
  - cursor: updated_at
  - fields: contacts_count(), created_at(), id(), name(), query_version(), sample_updated_at(), updated_at()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), email(), first_name(), id(), last_name(), phone_number(), updated_at()
- suppression_bounces:
  - primary key: email
  - cursor: created
  - fields: created(), email(), reason(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external SendGrid API read of marketing list, segment, contact, and suppression-bounce data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sendgrid
```

### Inspect as structured JSON

```bash
pm connectors inspect sendgrid --json
```

## Agent Rules

- Run pm connectors inspect sendgrid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
