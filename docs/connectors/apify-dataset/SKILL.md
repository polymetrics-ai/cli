---
name: pm-apify-dataset
description: Apify Dataset connector knowledge and safe action guide.
---

# pm-apify-dataset

## Purpose

Reads Apify dataset items and dataset metadata (item_collection, dataset_collection, dataset) through the Apify API v2. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/apify.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.apify.com/api/v2

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- dataset_id
- mode
- token (secret)

## ETL Streams

- item_collection:
  - fields: data()
- dataset_collection:
  - primary key: id
  - cursor: createdAt
  - fields: accessedAt(), actId(), actRunId(), cleanItemCount(), createdAt(), id(), itemCount(), modifiedAt(), name(), userId()
- dataset:
  - primary key: id
  - cursor: modifiedAt
  - fields: accessedAt(), actId(), actRunId(), cleanItemCount(), createdAt(), id(), itemCount(), modifiedAt(), name(), userId()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Apify Dataset API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect apify-dataset
```

### Inspect as structured JSON

```bash
pm connectors inspect apify-dataset --json
```

## Agent Rules

- Run pm connectors inspect apify-dataset before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
