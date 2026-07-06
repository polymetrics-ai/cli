---
name: pm-searxng
description: SearXNG connector knowledge and safe action guide.
---

# pm-searxng

## Purpose

Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json). Read-only. Requires base_url; no credentials by default.

## Icon

- asset: icons/searxng.svg
- source: official_site
- review_status: manual_override
- review_url: https://docs.searxng.org/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- query
- api_key (secret)

## ETL Streams

- search:
  - primary key: url
  - cursor: published_date
  - fields: category(), content(), engine(), engines(), published_date(), score(), stream(), thumbnail(), title(), url()
- reddit:
  - primary key: url
  - cursor: published_date
  - fields: category(), content(), engine(), engines(), published_date(), score(), stream(), thumbnail(), title(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external SearXNG instance read of web/Reddit search results
- approval: none; read-only public search API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect searxng
```

### Inspect as structured JSON

```bash
pm connectors inspect searxng --json
```

## Agent Rules

- Run pm connectors inspect searxng before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
