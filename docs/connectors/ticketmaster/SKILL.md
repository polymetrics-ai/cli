---
name: pm-ticketmaster
description: Ticketmaster connector knowledge and safe action guide.
---

# pm-ticketmaster

## Purpose

Reads events, venues, attractions, and classifications from the Ticketmaster Discovery API.

## Icon

- id: ticketmaster
- asset: icons/ticketmaster.svg
- source: official
- review_status: official_verified
- review_url: https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- country_code
- keyword
- locale
- api_key (secret)

## ETL Streams

- events:
  - primary key: id
  - fields: id(string), locale(string), name(string), type(string), url(string)
- venues:
  - primary key: id
  - fields: city(object), country(object), id(string), name(string), url(string)
- attractions:
  - primary key: id
  - fields: id(string), locale(string), name(string), type(string), url(string)
- classifications:
  - primary key: id
  - fields: genre(object), id(string), name(string), segment(object), subGenre(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Ticketmaster Discovery API read of public event/venue data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Ticketmaster's declared streams and reverse-ETL actions.
- Usage: pm ticketmaster <command> [flags]
- Read streams
- Other Commands
  - api get discovery v2 attractions - Documented GET /discovery/v2/attractions (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-attractions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 attractions id - Documented GET /discovery/v2/attractions/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-attractions-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 classifications - Documented GET /discovery/v2/classifications (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-classifications]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 classifications genres id - Documented GET /discovery/v2/classifications/genres/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-classifications-genres-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 classifications id - Documented GET /discovery/v2/classifications/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-classifications-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 classifications segments id - Documented GET /discovery/v2/classifications/segments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-classifications-segments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 classifications subgenres id - Documented GET /discovery/v2/classifications/subgenres/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-classifications-subgenres-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 events - Documented GET /discovery/v2/events (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 events id - Documented GET /discovery/v2/events/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-events-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 events id images - Documented GET /discovery/v2/events/{id}/images (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-events-id-images]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 suggest - Documented GET /discovery/v2/suggest (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-suggest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 venues - Documented GET /discovery/v2/venues (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-venues]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get discovery v2 venues id - Documented GET /discovery/v2/venues/{id} (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.discovery-v2-venues-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get events id json - Documented GET /events/{id}.json (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.events-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get suggest - Documented GET /suggest (not implemented) [intent=direct_read availability=not_implemented operation=ticketmaster.get.suggest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - attractions list - Run the attractions ETL stream [intent=etl availability=implemented stream=attractions]; notes: discrepancy=present-in-surface-absent-from-artifact
  - classifications list - Run the classifications ETL stream [intent=etl availability=implemented stream=classifications]; notes: discrepancy=present-in-surface-absent-from-artifact
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]; notes: discrepancy=present-in-surface-absent-from-artifact
  - venues list - Run the venues ETL stream [intent=etl availability=implemented stream=venues]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect ticketmaster
```

### Inspect as structured JSON

```bash
pm connectors inspect ticketmaster --json
```

## Agent Rules

- Run pm connectors inspect ticketmaster before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
