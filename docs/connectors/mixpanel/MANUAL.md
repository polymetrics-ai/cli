# pm connectors inspect mixpanel

```text
NAME
  pm connectors inspect mixpanel - Mixpanel connector manual

SYNOPSIS
  pm connectors inspect mixpanel
  pm connectors inspect mixpanel --json
  pm credentials add <name> --connector mixpanel [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mixpanel legacy Query API cohorts, annotations, engage profiles, and selected current Query/Annotations API list/detail endpoints.

ICON
  id: mixpanel
  asset: icons/mixpanel.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.mixpanel.com/reference/overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  analysis_type
  annotation_id
  base_url
  distinct_ids
  event_name
  from_date
  limit
  max_pages
  mode
  page_size
  project_id
  to_date
  username
  workspace_id
  api_secret (secret)
  password (secret)
  username_secret (secret)

ETL STREAMS
  cohorts:
    primary key: id
    fields: count(integer), id(integer), name(string)
  annotations:
    primary key: id
    fields: date(string), description(string), id(integer)
  engage:
    primary key: distinct_id
    fields: created(string), distinct_id(string), email(string)
  saved_funnels:
    primary key: funnel_id
    fields: funnel_id(integer), name(string)
  activity_stream:
    fields: event(string), properties(object)
  top_events:
    primary key: event
    fields: amount(integer), event(string), percent_change(number)
  event_property_names:
    primary key: name
    fields: count(integer), name(string)
  project_annotations:
    primary key: id
    fields: date(string), description(string), id(integer), tags(array), user(object)
  project_annotation:
    primary key: id
    fields: date(string), description(string), id(integer), tags(array), user(object)
  annotation_tags:
    primary key: id
    fields: has_annotations(boolean), id(integer), name(string), project_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Mixpanel Query/Application API read of cohort, annotation, profile, saved funnel, event breakdown, and annotation metadata
  approval: none; read-only Query API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mixpanel

  # Inspect as structured JSON
  pm connectors inspect mixpanel --json

AGENT WORKFLOW
  - Run pm connectors inspect mixpanel before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
