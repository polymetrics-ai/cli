---
name: pm-fastly
description: Fastly connector knowledge and safe action guide.
---

# pm-fastly

## Purpose

Reads Fastly services, the current user, the current customer (account), and datacenters through the Fastly REST API. Read-only.

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
- fastly_api_token (secret)

## ETL Streams

- services:
  - primary key: id
  - cursor: updated_at
  - fields: comment(), created_at(), customer_id(), deleted_at(), id(), name(), paused(), type(), updated_at(), version()
- current_user:
  - primary key: id
  - fields: created_at(), customer_id(), email_hash(), id(), locked(), login(), name(), role(), two_factor_auth_enabled(), updated_at()
- current_customer:
  - primary key: id
  - fields: billing_contact_id(), can_stream_syslog(), created_at(), has_account_panel(), id(), name(), owner_id(), pricing_plan(), updated_at()
- datacenters:
  - primary key: code
  - fields: code(), coordinates(), group(), name(), shield()
- service_details:
  - primary key: service_id
  - fields: activated_version(), comment(), created_at(), customer_id(), deleted_at(), environments(), id(), name(), service_id(), type(), updated_at(), version(), versions()
- users:
  - primary key: id
  - fields: created_at(), customer_id(), email_hash(), id(), locked(), login(), name(), role(), two_factor_auth_enabled(), updated_at()

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
