---
name: pm-newsdata
description: Newsdata connector knowledge and safe action guide.
---

# pm-newsdata

## Purpose

Reads latest news, cryptocurrency news, and news sources from the NewsData.io REST API.

## Icon

- id: source-newsdata
- asset: icons/source-newsdata.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://newsdata.io/documentation

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- category
- country
- domain
- language
- query
- query_in_title
- size
- api_key (secret) (required)

## ETL Streams

- latest:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_priority(integer), title(string)
- crypto:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_priority(integer), title(string)
- sources:
  - primary key: id
  - fields: category(array), country(array), description(string), icon(string), id(string), language(array), name(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NewsData.io API read of article and source metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect newsdata
```

### Inspect as structured JSON

```bash
pm connectors inspect newsdata --json
```

## Agent Rules

- Run pm connectors inspect newsdata before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
