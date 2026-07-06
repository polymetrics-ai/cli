---
name: pm-nytimes
description: New York Times connector knowledge and safe action guide.
---

# pm-nytimes

## Purpose

Reads New York Times Most Popular (viewed, emailed, shared) articles via the NYTimes Developer APIs.

## Icon

- asset: icons/nytimes.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.nytimes.com/apis

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- period
- api_key (secret)

## ETL Streams

- most_popular_viewed:
  - primary key: id
  - cursor: published_date
  - fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()
- most_popular_emailed:
  - primary key: id
  - cursor: published_date
  - fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()
- most_popular_shared:
  - primary key: id
  - cursor: published_date
  - fields: abstract(), byline(), id(), published_date(), section(), source(), title(), type(), updated(), uri(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NYTimes API read of published article metadata (no PII)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nytimes
```

### Inspect as structured JSON

```bash
pm connectors inspect nytimes --json
```

## Agent Rules

- Run pm connectors inspect nytimes before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
