# pm connectors inspect jira

```text
NAME
  pm connectors inspect jira - Jira connector manual

SYNOPSIS
  pm connectors inspect jira
  pm connectors inspect jira --json
  pm credentials add <name> --connector jira [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth (email + API token). Read-only.

ICON
  asset: icons/jira.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/changelog/#

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  email
  api_token (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updated
    fields: assignee(), created(), id(), issuetype(), key(), priority(), project(), reporter(), self(), status(), summary(), updated()
  projects:
    primary key: id
    fields: id(), isPrivate(), key(), name(), projectTypeKey(), self(), simplified(), style()
  users:
    primary key: accountId
    fields: accountId(), accountType(), active(), displayName(), emailAddress(), self()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Jira Cloud API read of issue, project, and user data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jira

  # Inspect as structured JSON
  pm connectors inspect jira --json

AGENT WORKFLOW
  - Run pm connectors inspect jira before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
