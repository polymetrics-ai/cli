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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  auth_token (secret)

ETL STREAMS
  organizations:
    primary key: id
    fields: api_key(), auto_upgrade(), collaborators_url(), created_at(), id(), name(), projects_url(), slug(), updated_at()
  projects:
    primary key: id
    fields: api_key(), collaborators_count(), created_at(), errors_url(), events_url(), for_review_error_count(), html_url(), id(), language(), name(), open_error_count(), organization_id(), slug(), type(), updated_at()
  collaborators:
    primary key: id
    fields: created_at(), email(), id(), is_admin(), last_request_at(), name(), pending_invitation(), two_factor_enabled()
  errors:
    primary key: id
    cursor: last_seen
    fields: comment_count(), context(), error_class(), events_count(), first_seen(), id(), last_seen(), message(), original_severity(), project_id(), severity(), status(), url()
  events:
    primary key: id
    cursor: received_at
    fields: context(), error_id(), id(), is_full_report(), project_id(), received_at(), severity(), unhandled(), url()
  releases:
    primary key: id
    cursor: release_time
    fields: app_bundle_version(), app_version(), app_version_code(), build_label(), errors_introduced_count(), errors_seen_count(), id(), project_id(), release_group_id(), release_source(), release_stage(), release_time()

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
