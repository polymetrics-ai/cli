---
name: pm-tempo
description: Tempo connector knowledge and safe action guide.
---

# pm-tempo

## Purpose

Reads Tempo accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4.

## Icon

- asset: icons/tempo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.tempo.io/

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
- api_token (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: global(), id(), key(), monthly_budget(), name(), status()
- customers:
  - primary key: id
  - fields: id(), key(), name()
- worklogs:
  - primary key: tempo_worklog_id
  - cursor: updated_at
  - fields: billable_seconds(), created_at(), description(), issue_id(), jira_worklog_id(), start_date(), start_time(), tempo_worklog_id(), time_spent_seconds(), updated_at()
- workload_schemes:
  - primary key: id
  - fields: default_scheme(), description(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Tempo Cloud API read of account, customer, and worklog data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tempo
```

### Inspect as structured JSON

```bash
pm connectors inspect tempo --json
```

## Agent Rules

- Run pm connectors inspect tempo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
