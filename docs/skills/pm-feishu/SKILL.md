---
name: pm-feishu
description: Feishu / Lark connector knowledge and safe action guide.
---

# pm-feishu

## Purpose

Reads Feishu/Lark Bitable (Base) records, tables, and field schemas via the Open Platform REST API using a tenant_access_token exchange. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: feishu
- asset: icons/feishu.svg
- source: official
- review_status: official_verified
- review_url: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- lark_host (required)
- mode
- page_size
- table_id (required)
- app_id (secret) (required)
- app_secret (secret) (required)
- app_token (secret) (required)

## ETL Streams

- records:
  - primary key: record_id
  - fields: fields(object), record_id(string)
- tables:
  - primary key: table_id
  - fields: name(string), revision(integer), table_id(string)
- fields:
  - primary key: field_id
  - fields: field_id(string), field_name(string), is_hidden(boolean), is_primary(boolean), property(object), type(integer), ui_type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Feishu / Lark API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect feishu
```

### Inspect as structured JSON

```bash
pm connectors inspect feishu --json
```

## Agent Rules

- Run pm connectors inspect feishu before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
