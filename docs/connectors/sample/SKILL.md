---
name: pm-sample
description: Sample connector knowledge and safe action guide.
---

# pm-sample

## Purpose

Built-in deterministic source connector for local development and tests.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- No connector-specific config fields.

## ETL Streams

- customers: Sample customer records.
  - primary key: id
  - cursor: updated_at
  - fields: id(string), name(string), email(string), plan(string), updated_at(timestamp)
- events: Sample event records.
  - primary key: id
  - cursor: occurred_at
  - fields: id(string), customer_id(string), event(string), occurred_at(timestamp)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
- Source modes: full_refresh, incremental
- Destination modes: append, overwrite, append_dedup, overwrite_dedup

## Security

- read risk: local deterministic sample data
- write risk: unsupported
- mutation risk: none
- approval: not required for reads
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sample
```

### Inspect as structured JSON

```bash
pm connectors inspect sample --json
```

### Sample ETL

```bash
pm credentials add sample-local --connector sample
pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers
pm etl run --connection sample_to_warehouse --stream customers --json
```

## Agent Rules

- Run pm connectors inspect sample before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
