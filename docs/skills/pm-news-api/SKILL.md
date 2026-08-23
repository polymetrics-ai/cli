---
name: pm-news-api
description: News API connector knowledge and safe action guide.
---

# pm-news-api

## Purpose

Reads articles and news sources from the News API (newsapi.org): the everything search, top headlines, and the sources directory.

## Icon

- id: newsapi
- asset: icons/newsapi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://newsapi.org/docs

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- category
- country
- domains
- end_date
- exclude_domains
- language
- search_in
- search_query
- sort_by
- sources
- start_date
- api_key (secret) (required)

## ETL Streams

- everything:
  - primary key: url
  - cursor: published_at
  - fields: author(string), content(string), description(string), published_at(string), source_id(string), source_name(string), title(string), url(string), url_to_image(string)
- top_headlines:
  - primary key: url
  - cursor: published_at
  - fields: author(string), content(string), description(string), published_at(string), source_id(string), source_name(string), title(string), url(string), url_to_image(string)
- sources:
  - primary key: id
  - fields: category(string), country(string), description(string), id(string), language(string), name(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external News API read of article and source metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect news-api
```

### Inspect as structured JSON

```bash
pm connectors inspect news-api --json
```

## Agent Rules

- Run pm connectors inspect news-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
