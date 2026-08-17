# pm connectors inspect amplitude

```text
NAME
  pm connectors inspect amplitude - Amplitude connector manual

SYNOPSIS
  pm connectors inspect amplitude
  pm connectors inspect amplitude --json
  pm credentials add <name> --connector amplitude [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and manages Amplitude behavioral cohorts, chart annotations, annotation categories, event lists, and the governed taxonomy (event/category definitions) through the Amplitude Analytics REST API.

ICON
  asset: icons/amplitude.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  taxonomy_show_deleted
  api_key (secret)
  secret_key (secret)

ETL STREAMS
  cohorts:
    primary key: id
    fields: archived(), createdAt(), description(), id(), lastComputed(), lastMod(), name(), owners(), published(), size(), type()
  cohorts_usage:
    primary key: resets_at
    fields: limit(), resets_at(), usage()
  annotations:
    primary key: id
    fields: date(), details(), id(), label()
  annotation_categories:
    primary key: id
    fields: id(), name()
  events_list:
    primary key: value
    fields: deleted(), display(), flow_hidden(), hidden(), non_active(), totals(), value()
  taxonomy_categories:
    primary key: id
    fields: id(), name()
  taxonomy_events:
    primary key: event_type
    fields: category(), description(), display_name(), event_type(), is_active(), is_hidden_from_dropdowns(), is_hidden_from_pathfinder(), is_hidden_from_persona_results(), is_hidden_from_timeline(), owner(), tags()
  taxonomy_event_properties:
    primary key: event_property, event_type
    fields: classifications(), description(), enum_values(), event_property(), event_type(), is_array_type(), is_hidden(), is_required(), regex(), type()
  taxonomy_user_properties:
    primary key: user_property
    fields: classifications(), deleted(), description(), enum_values(), is_array_type(), is_hidden(), regex(), type(), user_property()
  taxonomy_group_properties:
    primary key: group_type, group_property
    fields: classifications(), description(), enum_values(), group_property(), group_type(), is_array_type(), is_hidden(), regex(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_annotation:
    endpoint: POST /api/3/annotations
    risk: creates a chart annotation visible to every Amplitude project user
  update_annotation:
    endpoint: PUT /api/3/annotations/{{ record.id }}
    required fields: id
    risk: mutates an existing chart annotation visible to every Amplitude project user
  delete_annotation:
    endpoint: DELETE /api/3/annotations/{{ record.id }}
    required fields: id
    risk: permanently deletes a chart annotation
  create_annotation_category:
    endpoint: POST /api/3/annotation-categories
    risk: creates a new annotation category shared across the Amplitude project
  update_annotation_category:
    endpoint: PUT /api/3/annotation-categories/{{ record.id }}
    required fields: id
    risk: renames an existing annotation category shared across the Amplitude project
  delete_annotation_category:
    endpoint: DELETE /api/3/annotation-categories/{{ record.id }}
    required fields: id
    risk: permanently deletes an annotation category shared across the Amplitude project
  create_taxonomy_category:
    endpoint: POST /api/2/taxonomy/category
    risk: creates a new event category in the Amplitude project's governed taxonomy
  update_taxonomy_category:
    endpoint: PUT /api/2/taxonomy/category/{{ record.category_id }}
    required fields: category_id
    risk: renames an existing event category in the Amplitude project's governed taxonomy
  delete_taxonomy_category:
    endpoint: DELETE /api/2/taxonomy/category/{{ record.category_id }}
    required fields: category_id
    risk: permanently deletes an event category from the Amplitude project's governed taxonomy
  create_taxonomy_event:
    endpoint: POST /api/2/taxonomy/event
    risk: registers a new governed event type in the Amplitude project's taxonomy
  update_taxonomy_event:
    endpoint: PUT /api/2/taxonomy/event/{{ record.event_type }}
    required fields: event_type
    risk: mutates an existing governed event type's taxonomy metadata
  delete_taxonomy_event:
    endpoint: DELETE /api/2/taxonomy/event/{{ record.event_type }}
    required fields: event_type
    risk: soft-deletes a governed event type from the Amplitude project's taxonomy (recoverable via the restore endpoint, not modeled as a separate write action)

SECURITY
  read risk: external Amplitude API read of behavioral analytics data
  write risk: external Amplitude API mutation of chart annotations, annotation categories, and governed taxonomy event/category definitions — never behavioral event data itself
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amplitude

  # Inspect as structured JSON
  pm connectors inspect amplitude --json

AGENT WORKFLOW
  - Run pm connectors inspect amplitude before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
