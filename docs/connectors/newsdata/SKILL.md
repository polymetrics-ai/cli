---
name: pm-newsdata
description: Newsdata connector knowledge and safe action guide.
---

# pm-newsdata

## Purpose

Reads latest news, cryptocurrency news, and news sources from the NewsData.io REST API.

## Icon

- id: source-newsdata
- asset: icons/source-newsdata.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://newsdata.io/documentation

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- category
- country
- domain
- language
- query
- query_in_title
- size
- api_key (secret)

## ETL Streams

- latest:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_priority(integer), title(string)
- crypto:
  - primary key: article_id
  - cursor: pubDate
  - fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_priority(integer), title(string)
- sources:
  - primary key: id
  - fields: category(array), country(array), description(string), icon(string), id(string), language(array), name(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NewsData.io API read of article and source metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Newsdata's declared streams and reverse-ETL actions.
- Usage: pm newsdata <command> [flags]
- Read streams
- Other Commands
  - api get 1 archive - Documented GET /1/archive (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-archive]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get 1 count - Documented GET /1/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get 1 crypto count - Documented GET /1/crypto/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-crypto-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get 1 market - Documented GET /1/market (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-market]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get 1 market count - Documented GET /1/market/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-market-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get 1 news - Documented GET /1/news (not implemented) [intent=direct_read availability=not_implemented operation=newsdata.get.1-news]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - crypto list - Run the crypto ETL stream [intent=etl availability=implemented stream=crypto]
  - latest list - Run the latest ETL stream [intent=etl availability=implemented stream=latest]
  - sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]

## Commands

### Inspect as a manual

```bash
pm connectors inspect newsdata
```

### Inspect as structured JSON

```bash
pm connectors inspect newsdata --json
```

## Agent Rules

- Run pm connectors inspect newsdata before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
