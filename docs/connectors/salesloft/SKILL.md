---
name: pm-salesloft
description: Salesloft connector knowledge and safe action guide.
---

# pm-salesloft

## Purpose

Reads Salesloft people, accounts, cadences, users, and emails through the Salesloft REST API v2.

## Icon

- asset: icons/salesloft.svg
- source: official
- review_status: official_verified
- review_url: https://developers.salesloft.com/docs/api/

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
- token_url
- access_token (secret)
- api_key (secret)
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- people:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(), created_at(), display_name(), do_not_contact(), email_address(), first_name(), id(), last_name(), owner_id(), person_company_name(), phone(), title(), updated_at()
- accounts:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(), city(), company_type(), country(), created_at(), domain(), id(), industry(), name(), owner_id(), phone(), updated_at(), website()
- cadences:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(), created_at(), id(), name(), remove_bounces_enabled(), remove_replies_enabled(), shared(), team_cadence(), updated_at()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: active(), created_at(), email(), first_name(), guid(), id(), last_name(), name(), time_zone(), updated_at()
- emails:
  - primary key: id
  - cursor: updated_at
  - fields: bounced(), click_tracking(), created_at(), id(), sent_at(), status(), subject(), updated_at(), view_tracking()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Salesloft API read of people, accounts, cadences, users, and email data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect salesloft
```

### Inspect as structured JSON

```bash
pm connectors inspect salesloft --json
```

## Agent Rules

- Run pm connectors inspect salesloft before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
