---
name: pm-gutendex
description: Gutendex connector knowledge and safe action guide.
---

# pm-gutendex

## Purpose

Reads Project Gutenberg books from the free, public Gutendex JSON API (books, popular, latest, and English-language views). Read-only; no credentials required.

## Icon

- id: source-gutendex
- asset: icons/source-gutendex.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- author_year_end
- author_year_start
- base_url
- copyright
- ids
- languages
- mode
- search
- sort
- topic

## ETL Streams

- books:
  - primary key: id
  - fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
- popular_books:
  - primary key: id
  - fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
- latest_books:
  - primary key: id
  - fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
- english_books:
  - primary key: id
  - fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external read of the public, unauthenticated Gutendex book catalog
- approval: none; read-only public API, no credentials
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gutendex
```

### Inspect as structured JSON

```bash
pm connectors inspect gutendex --json
```

## Agent Rules

- Run pm connectors inspect gutendex before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
