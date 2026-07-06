---
name: pm-giphy
description: Giphy connector knowledge and safe action guide.
---

# pm-giphy

## Purpose

Reads GIFs, stickers, and clips from the Giphy search and trending REST endpoints. Read-only.

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
- max_pages
- mode
- page_size
- query_for_clips
- query_for_gif
- query_for_stickers
- rating
- api_key (secret)

## ETL Streams

- gif_search:
  - primary key: id
  - fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
- sticker_search:
  - primary key: id
  - fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
- clip_search:
  - primary key: id
  - fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
- trending_gifs:
  - primary key: id
  - fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Giphy API read of public media search/trending results
- approval: none; read-only public media source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect giphy
```

### Inspect as structured JSON

```bash
pm connectors inspect giphy --json
```

## Agent Rules

- Run pm connectors inspect giphy before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
