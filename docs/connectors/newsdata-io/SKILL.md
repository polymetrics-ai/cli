---
name: pm-newsdata-io
description: NewsData.io connector knowledge and safe action guide.
---

# pm-newsdata-io

## Purpose

Reads latest, crypto, and archived news articles plus available news sources from the NewsData.io REST API.

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
- categories
- countries
- domains
- end_date
- languages
- mode
- page_size
- search_query
- start_date
- api_key (secret) (required)

## ETL Streams

- latest:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
- crypto:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
- archive:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
- sources:
  - primary key: id
  - fields: category(array), country(array), description(string), icon(string), id(string), language(array), name(string), priority(integer), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NewsData.io API read of news articles and sources
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect newsdata-io
```

### Inspect as structured JSON

```bash
pm connectors inspect newsdata-io --json
```

## Agent Rules

- Run pm connectors inspect newsdata-io before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
