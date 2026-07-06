---
name: pm-newsdata-io
description: NewsData.io connector knowledge and safe action guide.
---

# pm-newsdata-io

## Purpose

Reads latest, crypto, and archived news articles plus available news sources from the NewsData.io REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
- api_key (secret)

## ETL Streams

- latest:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
- crypto:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
- archive:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
- sources:
  - primary key: id
  - fields: category(), country(), description(), icon(), id(), language(), name(), priority(), url()

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
