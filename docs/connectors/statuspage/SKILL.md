---
name: pm-statuspage
description: Statuspage connector knowledge and safe action guide.
---

# pm-statuspage

## Purpose

Reads Statuspage pages, components, incidents, subscribers, component groups, metrics, metrics providers, page access groups/users, and incident templates through the Statuspage API.

## Icon

- id: statuspage
- asset: icons/statuspage.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.statuspage.io/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_id
- api_key (secret) (required)

## ETL Streams

- pages:
  - primary key: id
  - fields: id(string), name(string), url(string)
- components:
  - primary key: id
  - fields: created_at(string), id(string), name(string), status(string)
- incidents:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), status(string)
- subscribers:
  - primary key: id
  - fields: created_at(string), id(string), name(string), status(string)
- component_groups:
  - primary key: id
  - fields: created_at(string), description(string), id(string), name(string), page_id(string), position(integer), updated_at(string)
- metrics:
  - primary key: id
  - fields: backfilled(boolean), created_at(string), decimal_places(integer), display(boolean), id(string), last_fetched_at(string), metric_identifier(string), metrics_provider_id(string), most_recent_data_at(string), name(string), suffix(string), tooltip_description(string), updated_at(string)
- metrics_providers:
  - primary key: id
  - fields: created_at(string), disabled(boolean), id(string), last_revalidated_at(string), metric_base_uri(string), page_id(string), type(string), updated_at(string)
- page_access_groups:
  - primary key: id
  - fields: component_ids(array), created_at(string), external_identifier(string), id(string), metric_ids(array), name(string), page_access_user_ids(array), page_id(string), updated_at(string)
- page_access_users:
  - primary key: id
  - fields: created_at(string), email(string), external_login(string), id(string), page_access_group_id(string), page_access_group_ids(array), page_id(string), updated_at(string)
- incident_templates:
  - primary key: id
  - fields: body(string), group_id(string), id(string), name(string), should_send_notifications(boolean), should_tweet(boolean), title(string), update_status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Statuspage API read of page, component, incident, subscriber, component group, metric, metrics provider, page access group/user, and incident template data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect statuspage
```

### Inspect as structured JSON

```bash
pm connectors inspect statuspage --json
```

## Agent Rules

- Run pm connectors inspect statuspage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
