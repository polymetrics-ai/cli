---
name: pm-mailtrap
description: Mailtrap connector knowledge and safe action guide.
---

# pm-mailtrap

## Purpose

Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap account-management REST API.

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

- account_id
- base_url
- api_token (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: access_levels(), id(), name()
- inboxes:
  - primary key: id
  - fields: account_id(), domain(), email_username(), emails_count(), id(), max_size(), name(), status(), used_size()
- projects:
  - primary key: id
  - fields: account_id(), id(), name()
- sending_domains:
  - primary key: id
  - fields: account_id(), demo(), domain_name(), id(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mailtrap API read of account-management data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailtrap
```

### Inspect as structured JSON

```bash
pm connectors inspect mailtrap --json
```

## Agent Rules

- Run pm connectors inspect mailtrap before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
