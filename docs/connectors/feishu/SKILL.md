---
name: pm-feishu
description: Feishu / Lark connector knowledge and safe action guide.
---

# pm-feishu

## Purpose

Reads Feishu/Lark Bitable (Base) records, tables, and field schemas via the Open Platform REST API using a tenant_access_token exchange. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

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
- lark_host
- mode
- page_size
- table_id
- app_id (secret)
- app_secret (secret)
- app_token (secret)

## ETL Streams

- records:
  - primary key: record_id
  - fields: fields(), record_id()
- tables:
  - primary key: table_id
  - fields: name(), revision(), table_id()
- fields:
  - primary key: field_id
  - fields: field_id(), field_name(), is_hidden(), is_primary(), property(), type(), ui_type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
