---
name: pm-bugsnag
description: Bugsnag connector knowledge and safe action guide.
---

# pm-bugsnag

## Purpose

Reads Bugsnag organizations, projects, collaborators, errors, events, and releases through the Bugsnag Data Access API.

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
- organization_id
- page_size
- project_id
- auth_token (secret)

## ETL Streams

- organizations:
  - primary key: id
  - fields: api_key(), auto_upgrade(), collaborators_url(), created_at(), id(), name(), projects_url(), slug(), updated_at()
- projects:
  - primary key: id
  - fields: api_key(), collaborators_count(), created_at(), errors_url(), events_url(), for_review_error_count(), html_url(), id(), language(), name(), open_error_count(), organization_id(), slug(), type(), updated_at()
- collaborators:
  - primary key: id
  - fields: created_at(), email(), id(), is_admin(), last_request_at(), name(), pending_invitation(), two_factor_enabled()
- errors:
  - primary key: id
  - cursor: last_seen
  - fields: comment_count(), context(), error_class(), events_count(), first_seen(), id(), last_seen(), message(), original_severity(), project_id(), severity(), status(), url()
- events:
  - primary key: id
  - cursor: received_at
  - fields: context(), error_id(), id(), is_full_report(), project_id(), received_at(), severity(), unhandled(), url()
- releases:
  - primary key: id
  - cursor: release_time
  - fields: app_bundle_version(), app_version(), app_version_code(), build_label(), errors_introduced_count(), errors_seen_count(), id(), project_id(), release_group_id(), release_source(), release_stage(), release_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Bugsnag API read of organization, project, collaborator, and error/event/release data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bugsnag
```

### Inspect as structured JSON

```bash
pm connectors inspect bugsnag --json
```

## Agent Rules

- Run pm connectors inspect bugsnag before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
