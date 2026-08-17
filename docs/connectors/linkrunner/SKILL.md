---
name: pm-linkrunner
description: Linkrunner connector knowledge and safe action guide.
---

# pm-linkrunner

## Purpose

Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API.

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
- channel
- display_id
- end_timestamp
- filter
- max_pages
- mode
- page_size
- start_timestamp
- timezone
- linkrunner-key (secret)

## ETL Streams

- campaigns:
  - primary key: display_id
  - cursor: update_at
  - fields: active(), attributed_users(), created_at(), default_link(), display_id(), domain(), google(), link(), meta(), meta_campaign_id(), name(), shareable_link(), update_at(), website()
- attributed_users:
  - primary key: campaign_display_id, attributed_at
  - cursor: attributed_at
  - fields: ad_set_id(), attributed_at(), campaign_display_id(), campaign_name(), installed_at(), link(), store_click_at(), user_data()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Linkrunner API read of mobile attribution campaign and user data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect linkrunner
```

### Inspect as structured JSON

```bash
pm connectors inspect linkrunner --json
```

## Agent Rules

- Run pm connectors inspect linkrunner before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
