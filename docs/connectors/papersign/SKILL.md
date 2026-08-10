---
name: pm-papersign
description: PaperSign connector knowledge and safe action guide.
---

# pm-papersign

## Purpose

Reads PaperSign documents, templates, and recipients through the REST API.

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
- api_key (secret) (required)

## ETL Streams

- documents:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), status(string), updated_at(string)
- templates:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), updated_at(string)
- recipients:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), document_id(string), email(string), id(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external PaperSign API read of document, template, and recipient data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect papersign
```

### Inspect as structured JSON

```bash
pm connectors inspect papersign --json
```

## Agent Rules

- Run pm connectors inspect papersign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
