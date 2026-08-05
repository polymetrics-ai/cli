# pm connectors inspect gitlab

```text
NAME
  pm connectors inspect gitlab - GitLab connector manual

SYNOPSIS
  pm connectors inspect gitlab
  pm connectors inspect gitlab --json
  pm credentials add <name> --connector gitlab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GitLab projects, groups, users, and issues through existing GitLab REST API v4 ETL streams; its provider-owned ledger records the full published surface, but G1 declares no write command.

ICON
  id: gitlab
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
  write risk: GitLab publishes mutation endpoints, but this G1 bundle declares no executable write actions.
  approval: Any future GitLab write must use the standard plan, preview, explicit approval, and execute flow; no GitLab writes are executable in this wave.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read the four existing GitLab ETL streams through typed commands.
  Usage: pm gitlab <command> [flags]
  Source CLI: GitLab REST API v4 (Official OpenAPI 3.0.0, info.version 19.3.0-pre, retrieved 2026-08-05)
  Global flags:
    --credential (string): Credential name to use for the GitLab request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum records to emit from a stream command.
  ETL streams
    projects list - List GitLab projects as ETL records. [intent=etl availability=implemented stream=projects]
    groups list - List GitLab groups as ETL records. [intent=etl availability=implemented stream=groups]
    users list - List GitLab users as ETL records. [intent=etl availability=implemented stream=users]
    issues list - List GitLab issues as ETL records. [intent=etl availability=implemented stream=issues]
  Help topics:
    authentication - Store GitLab access tokens with pm credentials; never pass secret values in command text.
    operation-inventory - The provider-owned GitLab OpenAPI v3 ledger records all 1,745 callable operations; this wave exposes only the four existing stream reads.
    known-limits - Of 1,745 inventoried operations, 4 are executable here; 1,618 need connector declarations, 45 need the multipart/file-upload foundation, 64 are provider-restricted, and 14 are deprecated exclusions.

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
