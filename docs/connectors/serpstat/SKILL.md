---
name: pm-serpstat
description: Serpstat connector knowledge and safe action guide.
---

# pm-serpstat

## Purpose

Reads Serpstat SEO domain keyword, competitor, and top-URL data through the Serpstat JSON-RPC-over-HTTP API. Read-only.

## Icon

- asset: icons/serpstat.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- domain
- page_size
- pages_to_fetch
- region_id
- api_key (secret)

## ETL Streams

- domain_keywords:
  - primary key: keyword, url
  - fields: keyword(), position(), updated_at(), url()
- domain_competitors:
  - primary key: domain
  - fields: domain(), visibility()
- domain_urls:
  - primary key: url
  - fields: keywords(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Serpstat API read of domain keyword/competitor/top-URL SEO metrics
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect serpstat
```

### Inspect as structured JSON

```bash
pm connectors inspect serpstat --json
```

## Agent Rules

- Run pm connectors inspect serpstat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
