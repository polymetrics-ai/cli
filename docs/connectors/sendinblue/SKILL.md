---
name: pm-sendinblue
description: Sendinblue connector knowledge and safe action guide.
---

# pm-sendinblue

## Purpose

Reads Sendinblue/Brevo contacts, campaigns, lists, and senders through the Brevo API.

## Icon

- asset: icons/sendinblue.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.brevo.com/reference

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: modifiedAt
  - fields: email(), id(), modifiedAt()
- email_campaigns:
  - primary key: id
  - cursor: modifiedAt
  - fields: id(), modifiedAt(), name(), status()
- contacts_lists:
  - primary key: id
  - fields: id(), name()
- senders:
  - primary key: id
  - fields: email(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Brevo (Sendinblue) API read of contact, campaign, list, and sender data
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sendinblue
```

### Inspect as structured JSON

```bash
pm connectors inspect sendinblue --json
```

## Agent Rules

- Run pm connectors inspect sendinblue before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
