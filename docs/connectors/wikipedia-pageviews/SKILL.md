---
name: pm-wikipedia-pageviews
description: Wikipedia Pageviews connector knowledge and safe action guide.
---

# pm-wikipedia-pageviews

## Purpose

Reads Wikimedia pageview metrics for articles and top-article reports through the public Wikimedia REST API.

## Icon

- id: wikipedia-pageviews
- asset: icons/wikipedia-pageviews.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- access (required)
- agent (required)
- article (required)
- base_url
- country
- day
- end (required)
- month
- project (required)
- start (required)
- year

## ETL Streams

- pageviews:
  - primary key: id
  - cursor: timestamp
  - fields: access(string), agent(string), article(string), granularity(string), id(string), project(string), timestamp(string), views(integer)
- top_articles:
  - primary key: id
  - fields: access(string), articles(array), country(string), day(string), id(string), month(string), project(string), year(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Wikimedia public API read of aggregate pageview metrics; no authentication, no PII
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect wikipedia-pageviews
```

### Inspect as structured JSON

```bash
pm connectors inspect wikipedia-pageviews --json
```

## Agent Rules

- Run pm connectors inspect wikipedia-pageviews before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
