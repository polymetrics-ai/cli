---
name: pm-giphy
description: Giphy connector knowledge and safe action guide.
---

# pm-giphy

## Purpose

Reads GIFs, stickers, and clips from the Giphy search and trending REST endpoints. Read-only.

## Icon

- id: simple-icons-giphy
- asset: icons/simple-icons/giphy.svg
- title: GIPHY
- simple_icon_slug: giphy
- simple_icon_hex: FF6666
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=GIPHY
- match: exact-name-or-slug
- matched_by: giphy

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
- api_key (secret) (required)

## ETL Streams

- gif_search:
  - primary key: id
  - fields: bitly_url(string), content_url(string), embed_url(string), id(string), import_datetime(string), rating(string), slug(string), source(string), source_tld(string), title(string), trending_datetime(string), type(string), url(string), username(string)
- sticker_search:
  - primary key: id
  - fields: bitly_url(string), content_url(string), embed_url(string), id(string), import_datetime(string), rating(string), slug(string), source(string), source_tld(string), title(string), trending_datetime(string), type(string), url(string), username(string)
- clip_search:
  - primary key: id
  - fields: bitly_url(string), content_url(string), embed_url(string), id(string), import_datetime(string), rating(string), slug(string), source(string), source_tld(string), title(string), trending_datetime(string), type(string), url(string), username(string)
- trending_gifs:
  - primary key: id
  - fields: bitly_url(string), content_url(string), embed_url(string), id(string), import_datetime(string), rating(string), slug(string), source(string), source_tld(string), title(string), trending_datetime(string), type(string), url(string), username(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
