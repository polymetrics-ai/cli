# Overview

Mixpanel reads 10 stream(s).

Readable streams: `cohorts`, `annotations`, `engage`, `saved_funnels`, `activity_stream`,
`top_events`, `event_property_names`, `project_annotations`, `project_annotation`,
`annotation_tags`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.mixpanel.com/reference/overview.

## Auth setup

Connection fields:

- `analysis_type` (optional, string); default `unique`; Mixpanel event analysis type for event
  breakdown streams: general, unique, or average.
- `annotation_id` (optional, string); Annotation id for the project_annotation detail stream.
- `api_secret` (optional, secret, string); Never logged.
- `base_url` (optional, string); default `https://mixpanel.com/api/2.0`; format `uri`; Mixpanel
  Query API base URL override for tests or proxies.
- `distinct_ids` (optional, string); JSON array string of distinct ids for the activity_stream Query
  API stream.
- `event_name` (optional, string); Event name for event-property breakdown streams.
- `from_date` (optional, string); Optional yyyy-mm-dd lower date bound for current Query or
  Annotations API streams.
- `limit` (optional, string); Optional result limit for Query API list-style streams.
- `max_pages` (optional, string); Maximum pages; leave unset, or use all/unlimited, to exhaust the
  stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `1000`; Records per page (1-10000).
- `password` (optional, secret, string); Mixpanel Query API service account secret (password). Used
  only for Basic auth; never logged.
- `project_id` (optional, string); Mixpanel project id used by current Query API, Annotations API,
  and other project-scoped API endpoints.
- `to_date` (optional, string); Optional yyyy-mm-dd upper date bound for current Query or
  Annotations API streams.
- `username` (optional, string).
- `username_secret` (optional, secret, string); Never logged.
- `workspace_id` (optional, string); Optional Mixpanel workspace id for Query API endpoints that
  accept workspace_id.

Secret fields are redacted in logs and write previews: `api_secret`, `password`, `username_secret`.

Default configuration values: `analysis_type=unique`, `base_url=https://mixpanel.com/api/2.0`,
`page_size=1000`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/cohorts/list`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next`.

- `cohorts`: GET `/cohorts/list` - records path `cohorts`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `page`; next token from `next`.
- `annotations`: GET `/annotations` - records path `annotations`; query `limit`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `page`; next token from `next`.
- `engage`: GET `/engage` - records path `results`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `next`.
- `saved_funnels`: GET `https://mixpanel.com/api/query/funnels/list` - records path `.`; query
  `project_id` from template `{{ config.project_id }}`, omitted when absent; `workspace_id` from
  template `{{ config.workspace_id }}`, omitted when absent; cursor pagination; cursor parameter
  `page`; next token from `next`.
- `activity_stream`: GET `https://mixpanel.com/api/query/stream/query` - records path
  `results.events`; query `distinct_ids`=`{{ config.distinct_ids }}`; `from_date`=`{{
  config.from_date }}`; `project_id` from template `{{ config.project_id }}`, omitted when absent;
  `to_date`=`{{ config.to_date }}`; `workspace_id` from template `{{ config.workspace_id }}`,
  omitted when absent; cursor pagination; cursor parameter `page`; next token from `next`.
- `top_events`: GET `https://mixpanel.com/api/query/events/top` - records path `events`; query
  `limit` from template `{{ config.limit }}`, omitted when absent; `project_id` from template `{{
  config.project_id }}`, omitted when absent; `type`=`{{ config.analysis_type }}`; `workspace_id`
  from template `{{ config.workspace_id }}`, omitted when absent; cursor pagination; cursor
  parameter `page`; next token from `next`.
- `event_property_names`: GET `https://mixpanel.com/api/query/events/properties/top` - records path
  `.`; flattens keyed objects; key field `name`; query `event`=`{{ config.event_name }}`; `limit`
  from template `{{ config.limit }}`, omitted when absent; `project_id` from template `{{
  config.project_id }}`, omitted when absent; `workspace_id` from template `{{ config.workspace_id
  }}`, omitted when absent; cursor pagination; cursor parameter `page`; next token from `next`.
- `project_annotations`: GET `https://mixpanel.com/api/app/projects/{{ config.project_id
  }}/annotations` - records path `results`; query `fromDate` from template `{{ config.from_date }}`,
  omitted when absent; `toDate` from template `{{ config.to_date }}`, omitted when absent; cursor
  pagination; cursor parameter `page`; next token from `next`.
- `project_annotation`: GET `https://mixpanel.com/api/app/projects/{{ config.project_id
  }}/annotations/{{ config.annotation_id }}` - single-object response; records path `results`;
  cursor pagination; cursor parameter `page`; next token from `next`.
- `annotation_tags`: GET `https://mixpanel.com/api/app/projects/{{ config.project_id
  }}/annotations/tags` - records path `.`; cursor pagination; cursor parameter `page`; next token
  from `next`.

## Write actions & risks

This connector is read-only. Read behavior: external Mixpanel Query/Application API read of cohort,
annotation, profile, saved funnel, event breakdown, and annotation metadata.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 10 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=7, duplicate_of=1, non_data_endpoint=2, out_of_scope=1,
  requires_elevated_scope=25.
