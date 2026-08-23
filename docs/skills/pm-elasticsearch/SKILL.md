---
name: pm-elasticsearch
description: Elasticsearch connector knowledge and safe action guide.
---

# pm-elasticsearch

## Purpose

Reads Elasticsearch index metadata and documents through the REST API. Read-only.

## Icon

- id: elasticsearch
- asset: icons/elasticsearch.svg
- source: official
- review_status: official_verified
- review_url: https://www.elastic.co/docs/reference/elasticsearch

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- endpoint (required)
- index
- max_pages
- mode
- page_size
- username
- api_key_id (secret)
- api_key_secret (secret)
- password (secret)

## ETL Streams

- indices:
  - primary key: index
  - fields: docs.count(string), index(string)
- documents:
  - primary key: id
  - fields: id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Elasticsearch cluster read of index metadata and documents
- approval: none; read-only cluster access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect elasticsearch
```

### Inspect as structured JSON

```bash
pm connectors inspect elasticsearch --json
```

## Agent Rules

- Run pm connectors inspect elasticsearch before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
