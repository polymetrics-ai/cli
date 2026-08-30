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
  id: sentry
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
  auth_token (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), id(string), isBookmarked(boolean), isPublic(boolean), name(string), platform(string), slug(string), status(string)
  issues:
    primary key: id
    cursor: lastSeen
    fields: count(string), culprit(string), firstSeen(string), id(string), lastSeen(string), level(string), shortId(string), status(string), title(string), type(string), userCount(integer)
  events:
    primary key: id
    cursor: dateCreated
    fields: dateCreated(string), eventID(string), groupID(string), id(string), message(string), platform(string), title(string), type(string)
  releases:
    primary key: version
    cursor: dateCreated
    fields: dateCreated(string), dateReleased(string), ref(string), shortVersion(string), status(string), url(string), version(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Sentry API read of project, issue, event, and release data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Sentry command surface
  Usage: pm sentry <command>
  Seer
  Other Commands
    seer list-models - List the declared Sentry Seer models. [intent=direct_read availability=implemented operation=sentry.seer_models_list]; flags: --page, --page-cursor

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
