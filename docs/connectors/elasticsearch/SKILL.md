---
name: pm-elasticsearch
description: Elasticsearch connector knowledge and safe action guide.
---

# pm-elasticsearch

## Purpose

Reads Elasticsearch index metadata and documents through the REST API. Read-only.

## Icon

- asset: icons/elasticsearch.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- endpoint
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
  - fields: docs.count(), index()
- documents:
  - primary key: id
  - fields: id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
