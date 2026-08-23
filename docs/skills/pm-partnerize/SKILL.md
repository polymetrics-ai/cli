---
name: pm-partnerize
description: Partnerize connector knowledge and safe action guide.
---

# pm-partnerize

## Purpose

Reads Partnerize conversions, campaigns, and publishers through the REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- application_key (secret) (required)
- user_api_key (secret) (required)

## ETL Streams

- conversions:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), currency(string), id(string), status(string), value(number)
- campaigns:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), status(string)
- publishers:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Partnerize API read of conversion, campaign, and publisher data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect partnerize
```

### Inspect as structured JSON

```bash
pm connectors inspect partnerize --json
```

## Agent Rules

- Run pm connectors inspect partnerize before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
