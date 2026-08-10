# pm connectors inspect bugsnag

```text
NAME
  pm connectors inspect bugsnag - Bugsnag connector manual

SYNOPSIS
  pm connectors inspect bugsnag
  pm connectors inspect bugsnag --json
  pm credentials add <name> --connector bugsnag [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bugsnag organizations, projects, collaborators, errors, events, and releases through the Bugsnag Data Access API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  organization_id
  page_size
  project_id
  auth_token (secret) (required)

ETL STREAMS
  organizations:
    primary key: id
    fields: api_key(string), auto_upgrade(boolean), collaborators_url(string), created_at(string), id(string), name(string), projects_url(string), slug(string), updated_at(string)
  projects:
    primary key: id
    fields: api_key(string), collaborators_count(integer), created_at(string), errors_url(string), events_url(string), for_review_error_count(integer), html_url(string), id(string), language(string), name(string), open_error_count(integer), organization_id(string), slug(string), type(string), updated_at(string)
  collaborators:
    primary key: id
    fields: created_at(string), email(string), id(string), is_admin(boolean), last_request_at(string), name(string), pending_invitation(boolean), two_factor_enabled(boolean)
  errors:
    primary key: id
    cursor: last_seen
    fields: comment_count(integer), context(string), error_class(string), events_count(integer), first_seen(string), id(string), last_seen(string), message(string), original_severity(string), project_id(string), severity(string), status(string), url(string)
  events:
    primary key: id
    cursor: received_at
    fields: context(string), error_id(string), id(string), is_full_report(boolean), project_id(string), received_at(string), severity(string), unhandled(boolean), url(string)
  releases:
    primary key: id
    cursor: release_time
    fields: app_bundle_version(string), app_version(string), app_version_code(string), build_label(string), errors_introduced_count(integer), errors_seen_count(integer), id(string), project_id(string), release_group_id(string), release_source(string), release_stage(string), release_time(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Bugsnag API read of organization, project, collaborator, and error/event/release data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bugsnag

  # Inspect as structured JSON
  pm connectors inspect bugsnag --json

AGENT WORKFLOW
  - Run pm connectors inspect bugsnag before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
