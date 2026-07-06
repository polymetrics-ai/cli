# pm connectors inspect sentry

```text
NAME
  pm connectors inspect sentry - Sentry connector manual

SYNOPSIS
  pm connectors inspect sentry
  pm connectors inspect sentry --json
  pm credentials add <name> --connector sentry [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Sentry projects, issues, error events, and releases through the Sentry REST API (read-only).

ICON
  asset: icons/sentry.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.sentry.io/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  organization
  page_size
  project
  auth_token (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(), id(), isBookmarked(), isPublic(), name(), platform(), slug(), status()
  issues:
    primary key: id
    cursor: lastSeen
    fields: count(), culprit(), firstSeen(), id(), lastSeen(), level(), shortId(), status(), title(), type(), userCount()
  events:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(), eventID(), groupID(), id(), message(), platform(), title(), type()
  releases:
    primary key: version
    cursor: dateCreated
    fields: dateCreated(), dateReleased(), ref(), shortVersion(), status(), url(), version()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Sentry API read of project, issue, event, and release data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sentry

  # Inspect as structured JSON
  pm connectors inspect sentry --json

AGENT WORKFLOW
  - Run pm connectors inspect sentry before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
