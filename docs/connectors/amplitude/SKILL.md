---
name: pm-amplitude
description: Amplitude connector knowledge and safe action guide.
---

# pm-amplitude

## Purpose

Reads and manages Amplitude behavioral cohorts, chart annotations, annotation categories, event lists, and the governed taxonomy (event/category definitions) through the Amplitude Analytics REST API.

## Icon

- id: amplitude
- asset: icons/amplitude.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- taxonomy_show_deleted
- api_key (secret) (required)
- secret_key (secret) (required)

## ETL Streams

- cohorts:
  - primary key: id
  - fields: archived(boolean), createdAt(integer), description(string), id(string), lastComputed(integer), lastMod(integer), name(string), owners(array), published(boolean), size(integer), type(string)
- cohorts_usage:
  - primary key: resets_at
  - fields: limit(integer), resets_at(string), usage(integer)
- annotations:
  - primary key: id
  - fields: date(string), details(string), id(integer), label(string)
- annotation_categories:
  - primary key: id
  - fields: id(integer), name(string)
- events_list:
  - primary key: value
  - fields: deleted(boolean), display(string), flow_hidden(boolean), hidden(boolean), non_active(boolean), totals(integer), value(string)
- taxonomy_categories:
  - primary key: id
  - fields: id(integer), name(string)
- taxonomy_events:
  - primary key: event_type
  - fields: category(object), description(string), display_name(string), event_type(string), is_active(boolean), is_hidden_from_dropdowns(boolean), is_hidden_from_pathfinder(boolean), is_hidden_from_persona_results(boolean), is_hidden_from_timeline(boolean), owner(string), tags(array)
- taxonomy_event_properties:
  - primary key: event_property, event_type
  - fields: classifications(array), description(string), enum_values(array), event_property(string), event_type(string), is_array_type(boolean), is_hidden(boolean), is_required(boolean), regex(string), type(string)
- taxonomy_user_properties:
  - primary key: user_property
  - fields: classifications(array), deleted(boolean), description(string), enum_values(array), is_array_type(boolean), is_hidden(boolean), regex(string), type(string), user_property(string)
- taxonomy_group_properties:
  - primary key: group_type, group_property
  - fields: classifications(array), description(string), enum_values(array), group_property(string), group_type(string), is_array_type(boolean), is_hidden(boolean), regex(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_annotation:
  - endpoint: POST /api/3/annotations
  - required fields: label, start
  - risk: creates a chart annotation visible to every Amplitude project user
- update_annotation:
  - endpoint: PUT /api/3/annotations/{{ record.id }}
  - required fields: id
  - risk: mutates an existing chart annotation visible to every Amplitude project user
- delete_annotation:
  - endpoint: DELETE /api/3/annotations/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a chart annotation
- create_annotation_category:
  - endpoint: POST /api/3/annotation-categories
  - required fields: category
  - risk: creates a new annotation category shared across the Amplitude project
- update_annotation_category:
  - endpoint: PUT /api/3/annotation-categories/{{ record.id }}
  - required fields: id, category
  - risk: renames an existing annotation category shared across the Amplitude project
- delete_annotation_category:
  - endpoint: DELETE /api/3/annotation-categories/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an annotation category shared across the Amplitude project
- create_taxonomy_category:
  - endpoint: POST /api/2/taxonomy/category
  - required fields: category_name
  - risk: creates a new event category in the Amplitude project's governed taxonomy
- update_taxonomy_category:
  - endpoint: PUT /api/2/taxonomy/category/{{ record.category_id }}
  - required fields: category_id, category_name
  - risk: renames an existing event category in the Amplitude project's governed taxonomy
- delete_taxonomy_category:
  - endpoint: DELETE /api/2/taxonomy/category/{{ record.category_id }}
  - required fields: category_id
  - risk: permanently deletes an event category from the Amplitude project's governed taxonomy
- create_taxonomy_event:
  - endpoint: POST /api/2/taxonomy/event
  - required fields: event_type
  - risk: registers a new governed event type in the Amplitude project's taxonomy
- update_taxonomy_event:
  - endpoint: PUT /api/2/taxonomy/event/{{ record.event_type }}
  - required fields: event_type
  - risk: mutates an existing governed event type's taxonomy metadata
- delete_taxonomy_event:
  - endpoint: DELETE /api/2/taxonomy/event/{{ record.event_type }}
  - required fields: event_type
  - risk: soft-deletes a governed event type from the Amplitude project's taxonomy (recoverable via the restore endpoint, not modeled as a separate write action)

## Security

- read risk: external Amplitude API read of behavioral analytics data
- write risk: external Amplitude API mutation of chart annotations, annotation categories, and governed taxonomy event/category definitions — never behavioral event data itself
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect amplitude
```

### Inspect as structured JSON

```bash
pm connectors inspect amplitude --json
```

## Agent Rules

- Run pm connectors inspect amplitude before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
