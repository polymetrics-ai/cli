---
name: pm-mailtrap
description: Mailtrap connector knowledge and safe action guide.
---

# pm-mailtrap

## Purpose

Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap account-management REST API.

## Icon

- id: simple-icons-mailtrap
- asset: icons/simple-icons/mailtrap.svg
- title: Mailtrap
- simple_icon_slug: mailtrap
- simple_icon_hex: 22D172
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Mailtrap
- match: exact-name-or-slug
- matched_by: mailtrap

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- api_token (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - fields: access_levels(array), id(integer), name(string)
- inboxes:
  - primary key: id
  - fields: account_id(string), domain(string), email_username(string), emails_count(integer), id(integer), max_size(integer), name(string), status(string), used_size(integer)
- projects:
  - primary key: id
  - fields: account_id(string), id(integer), name(string)
- sending_domains:
  - primary key: id
  - fields: account_id(string), demo(boolean), domain_name(string), id(integer), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
