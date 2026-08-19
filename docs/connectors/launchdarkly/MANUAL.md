# pm connectors inspect launchdarkly

```text
NAME
  pm connectors inspect launchdarkly - LaunchDarkly connector manual

SYNOPSIS
  pm connectors inspect launchdarkly
  pm connectors inspect launchdarkly --json
  pm credentials add <name> --connector launchdarkly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads LaunchDarkly projects, members, audit log entries, feature flags, and environments through the LaunchDarkly REST API.

ICON
  id: launchdarkly
  asset: icons/launchdarkly.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.launchdarkly.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  project_key
  access_token (secret) (required)

ETL STREAMS
  projects:
    primary key: _id
    fields: _id(string), key(string), name(string), tags(array)
  members:
    primary key: _id
    fields: _id(string), _pendingInvite(boolean), email(string), firstName(string), lastName(string), role(string)
  auditlog:
    primary key: _id
    cursor: date
    fields: _id(string), date(integer), description(string), kind(string), name(string), shortDescription(string)
  flags:
    primary key: key
    fields: creationDate(integer), description(string), key(string), kind(string), name(string), tags(array), temporary(boolean)
  environments:
    primary key: _id
    fields: _id(string), color(string), defaultTtl(integer), key(string), name(string), tags(array)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external LaunchDarkly API read of project, membership, audit, and feature-flag configuration data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect launchdarkly

  # Inspect as structured JSON
  pm connectors inspect launchdarkly --json

AGENT WORKFLOW
  - Run pm connectors inspect launchdarkly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
