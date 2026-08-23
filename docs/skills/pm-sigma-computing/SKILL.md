---
name: pm-sigma-computing
description: Sigma Computing connector knowledge and safe action guide.
---

# pm-sigma-computing

## Purpose

Reads Sigma workbooks, datasets, teams, and members through the Sigma REST API.

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
- page_size
- access_token (secret)

## ETL Streams

- workbooks:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), updated_at(string)
- datasets:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), updated_at(string)
- teams:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), updated_at(string)
- members:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Sigma Computing API read of workbook, dataset, team, and member data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sigma-computing
```

### Inspect as structured JSON

```bash
pm connectors inspect sigma-computing --json
```

## Agent Rules

- Run pm connectors inspect sigma-computing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
