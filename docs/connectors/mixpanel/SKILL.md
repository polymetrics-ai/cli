---
name: pm-mixpanel
description: Mixpanel connector knowledge and safe action guide.
---

# pm-mixpanel

## Purpose

Reads Mixpanel legacy Query API cohorts, annotations, engage profiles, and selected current Query/Annotations API list/detail endpoints.

## Icon

- asset: icons/mixpanel.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.mixpanel.com/reference/overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- analysis_type
- annotation_id
- base_url
- distinct_ids
- event_name
- from_date
- limit
- max_pages
- mode
- page_size
- project_id
- to_date
- username
- workspace_id
- api_secret (secret)
- password (secret)
- username_secret (secret)

## ETL Streams

- cohorts:
  - primary key: id
  - fields: count(), id(), name()
- annotations:
  - primary key: id
  - fields: date(), description(), id()
- engage:
  - primary key: distinct_id
  - fields: created(), distinct_id(), email()
- saved_funnels:
  - primary key: funnel_id
  - fields: funnel_id(), name()
- activity_stream:
  - fields: event(), properties()
- top_events:
  - primary key: event
  - fields: amount(), event(), percent_change()
- event_property_names:
  - primary key: name
  - fields: count(), name()
- project_annotations:
  - primary key: id
  - fields: date(), description(), id(), tags(), user()
- project_annotation:
  - primary key: id
  - fields: date(), description(), id(), tags(), user()
- annotation_tags:
  - primary key: id
  - fields: has_annotations(), id(), name(), project_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mixpanel Query/Application API read of cohort, annotation, profile, saved funnel, event breakdown, and annotation metadata
- approval: none; read-only Query API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mixpanel
```

### Inspect as structured JSON

```bash
pm connectors inspect mixpanel --json
```

## Agent Rules

- Run pm connectors inspect mixpanel before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
