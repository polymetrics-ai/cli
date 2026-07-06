---
name: pm-us-census
description: US Census connector knowledge and safe action guide.
---

# pm-us-census

## Purpose

Reads configured datasets from the US Census Bureau's API via a caller-supplied query path and query-string qualifier, and reads the Bureau's own published dataset catalog.

## Icon

- asset: icons/uscensus.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.census.gov/data/developers/data-sets.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- query_params
- query_path
- api_key (secret)

## ETL Streams

- query:
  - primary key: name
  - fields: estab(), name()
- datasets:
  - primary key: identifier
  - fields: accessLevel(), c_dataset(), c_geographyLink(), c_isAvailable(), c_variablesLink(), c_vintage(), dataset_path(), description(), identifier(), modified(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external US Census Bureau API read of a caller-configured dataset endpoint, plus the Bureau's own public dataset catalog (no auth required for the catalog)
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect us-census
```

### Inspect as structured JSON

```bash
pm connectors inspect us-census --json
```

## Agent Rules

- Run pm connectors inspect us-census before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
