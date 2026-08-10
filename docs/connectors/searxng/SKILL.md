---
name: pm-searxng
description: SearXNG connector knowledge and safe action guide.
---

# pm-searxng

## Purpose

Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json). Read-only. Requires base_url; no credentials by default.

## Icon

- id: searxng
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
  - fields: category(string), content(string), engine(string), engines(string), published_date(string), score(number), stream(string), thumbnail(string), title(string), url(string)
- reddit:
  - primary key: url
  - cursor: published_date
  - fields: category(string), content(string), engine(string), engines(string), published_date(string), score(number), stream(string), thumbnail(string), title(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external SearXNG instance read of web/Reddit search results
- approval: none; read-only public search API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run SearXNG's declared streams and reverse-ETL actions.
- Usage: pm searxng <command> [flags]
- Read streams
- Other Commands
  - api get config - Documented GET /config (not implemented) [intent=direct_read availability=not_implemented operation=searxng.get.config]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get healthz - Documented GET /healthz (not implemented) [intent=direct_read availability=not_implemented operation=searxng.get.healthz]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - reddit list - Run the reddit ETL stream [intent=etl availability=implemented stream=reddit]
  - search list - Run the search ETL stream [intent=etl availability=implemented stream=search]

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
