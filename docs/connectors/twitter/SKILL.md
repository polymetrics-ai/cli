---
name: pm-twitter
description: Twitter connector knowledge and safe action guide.
---

# pm-twitter

## Purpose

Reads tweets and their authors matching a search query from the Twitter (X) API v2 recent search endpoint using an App-only Bearer token.

## Icon

- asset: icons/twitter.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.twitter.com/en/docs/twitter-api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- end_date
- max_pages
- mode
- page_size
- query
- start_date
- api_key (secret)

## ETL Streams

- tweets:
  - primary key: id
  - cursor: created_at
  - fields: author_id(), conversation_id(), created_at(), id(), in_reply_to_user_id(), lang(), possibly_sensitive(), public_metrics(), source(), text()
- authors:
  - primary key: id
  - fields: created_at(), description(), id(), location(), name(), protected(), public_metrics(), url(), username(), verified()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Twitter (X) API read of tweets and author profiles matching a search query
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect twitter
```

### Inspect as structured JSON

```bash
pm connectors inspect twitter --json
```

## Agent Rules

- Run pm connectors inspect twitter before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
