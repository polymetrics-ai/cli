---
name: pm-nexus-datasets
description: Infor Nexus Datasets connector knowledge and safe action guide.
---

# pm-nexus-datasets

## Purpose

Reads records from a configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1) using HMAC-SHA256 request signing. Read-only.

## Icon

- id: nexus-datasets
- asset: icons/nexus-datasets.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- dataset_name (required)
- mode
- start_date
- access_key_id (secret) (required)
- api_key (secret) (required)
- secret_key (secret) (required)
- user_id (secret) (required)

## ETL Streams

- datasets:
  - primary key: id
  - cursor: updated_at
  - fields: dataset_name(string), id(string), raw_data(object), raw_data_string(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Infor Nexus dataset export read, HMAC-signed
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nexus-datasets
```

### Inspect as structured JSON

```bash
pm connectors inspect nexus-datasets --json
```

## Agent Rules

- Run pm connectors inspect nexus-datasets before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
