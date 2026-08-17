---
name: pm-gnews
description: GNews connector knowledge and safe action guide.
---

# pm-gnews

## Purpose

Reads GNews articles from the keyword search and top-headlines endpoints of the GNews REST API. Read-only.

## Icon

- asset: icons/gnews.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://gnews.io/docs/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- country
- end_date
- in
- language
- max_pages
- mode
- nullable
- page_size
- query
- sortby
- start_date
- top_headlines_query
- top_headlines_topic
- api_key (secret)

## ETL Streams

- search:
  - primary key: id
  - cursor: published_at
  - fields: content(), description(), id(), image(), lang(), published_at(), source_country(), source_id(), source_name(), source_url(), title(), url()
- top_headlines:
  - primary key: id
  - cursor: published_at
  - fields: content(), description(), id(), image(), lang(), published_at(), source_country(), source_id(), source_name(), source_url(), title(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external GNews API read of news article search results
- approval: none; read-only news search API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gnews
```

### Inspect as structured JSON

```bash
pm connectors inspect gnews --json
```

## Agent Rules

- Run pm connectors inspect gnews before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
