---
name: pm-nasa
description: NASA connector knowledge and safe action guide.
---

# pm-nasa

## Purpose

Reads NASA Open API data: Astronomy Picture of the Day, Near-Earth Objects (NeoWs feed and browse), EPIC Earth imagery, and Mars rover photos. Read-only.

## Icon

- asset: icons/nasa.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.nasa.gov/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- count
- end_date
- mode
- sol
- start_date
- thumbs
- api_key (secret)

## ETL Streams

- apod:
  - primary key: date
  - cursor: date
  - fields: copyright(), date(), explanation(), hdurl(), media_type(), service_version(), thumbnail_url(), title(), url()
- neo_feed:
  - primary key: id
  - fields: absolute_magnitude_h(), id(), is_potentially_hazardous_asteroid(), is_sentry_object(), name(), nasa_jpl_url(), neo_reference_id()
- neo_browse:
  - primary key: id
  - fields: absolute_magnitude_h(), id(), is_potentially_hazardous_asteroid(), is_sentry_object(), name(), nasa_jpl_url(), neo_reference_id()
- epic:
  - primary key: identifier
  - fields: caption(), date(), identifier(), image(), version()
- mars_photos:
  - primary key: id
  - fields: camera(), earth_date(), id(), img_src(), rover(), sol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NASA Open API read of public astronomy and space data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nasa
```

### Inspect as structured JSON

```bash
pm connectors inspect nasa --json
```

## Agent Rules

- Run pm connectors inspect nasa before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
