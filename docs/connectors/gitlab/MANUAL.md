# pm connectors inspect gitlab

```text
NAME
  pm connectors inspect gitlab - GitLab connector manual

SYNOPSIS
  pm connectors inspect gitlab
  pm connectors inspect gitlab --json
  pm credentials add <name> --connector gitlab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

ICON
  asset: icons/gitlab.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.gitlab.com/ee/api/rest/deprecations.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  start_date
  access_token (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: last_activity_at
    fields: archived(), created_at(), default_branch(), description(), forks_count(), id(), last_activity_at(), name(), open_issues_count(), path(), path_with_namespace(), star_count(), visibility(), web_url()
  groups:
    primary key: id
    cursor: created_at
    fields: created_at(), description(), full_name(), full_path(), id(), name(), parent_id(), path(), visibility(), web_url()
  users:
    primary key: id
    cursor: created_at
    fields: bot(), created_at(), id(), is_admin(), name(), state(), username(), web_url()
  issues:
    primary key: id
    cursor: updated_at
    fields: author_id(), closed_at(), created_at(), downvotes(), id(), iid(), project_id(), state(), title(), updated_at(), upvotes(), user_notes_count(), web_url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external GitLab API read of projects, groups, users, and issues
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gitlab

  # Inspect as structured JSON
  pm connectors inspect gitlab --json

AGENT WORKFLOW
  - Run pm connectors inspect gitlab before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
