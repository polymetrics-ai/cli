---
name: pm-fastly
description: Fastly connector knowledge and safe action guide.
---

# pm-fastly

## Purpose

Reads Fastly services, the current user, the current customer (account), and datacenters through the Fastly REST API. Read-only.

## Icon

- id: simple-icons-fastly
- asset: icons/simple-icons/fastly.svg
- title: Fastly
- simple_icon_slug: fastly
- simple_icon_hex: FF282D
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Fastly
- match: exact-name-or-slug
- matched_by: fastly

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- fastly_api_token (secret)

## ETL Streams

- services:
  - primary key: id
  - cursor: updated_at
  - fields: comment(string), created_at(string), customer_id(string), deleted_at(string), id(string), name(string), paused(boolean), type(string), updated_at(string), version(integer)
- current_user:
  - primary key: id
  - fields: created_at(string), customer_id(string), email_hash(string), id(string), locked(boolean), login(string), name(string), role(string), two_factor_auth_enabled(boolean), updated_at(string)
- current_customer:
  - primary key: id
  - fields: billing_contact_id(string), can_stream_syslog(boolean), created_at(string), has_account_panel(boolean), id(string), name(string), owner_id(string), pricing_plan(string), updated_at(string)
- datacenters:
  - primary key: code
  - fields: code(string), coordinates(object), group(string), name(string), shield(string)
- service_details:
  - primary key: service_id
  - fields: activated_version(object), comment(string), created_at(string), customer_id(string), deleted_at(string), environments(array), id(string), name(string), service_id(string), type(string), updated_at(string), version(object), versions(array)
- users:
  - primary key: id
  - fields: created_at(string), customer_id(string), email_hash(string), id(string), locked(boolean), login(string), name(string), role(string), two_factor_auth_enabled(boolean), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Fastly API read of service/account configuration metadata
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fastly
```

### Inspect as structured JSON

```bash
pm connectors inspect fastly --json
```

## Agent Rules

- Run pm connectors inspect fastly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
