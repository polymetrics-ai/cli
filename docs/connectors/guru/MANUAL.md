# pm connectors inspect guru

```text
NAME
  pm connectors inspect guru - Guru connector manual

SYNOPSIS
  pm connectors inspect guru
  pm connectors inspect guru --json
  pm credentials add <name> --connector guru [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Guru collections, groups, members, and teams through the Guru REST API using HTTP Basic authentication (email + API token).

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
  max_pages
  mode
  page_size
  username
  password (secret)

ETL STREAMS
  collections:
    primary key: id
    fields: collectionType(), color(), dateCreated(), description(), id(), name(), publicCardsEnabled(), slug()
  groups:
    primary key: id
    fields: dateCreated(), groupType(), id(), memberCount(), modifiable(), name()
  members:
    primary key: id
    fields: dateCreated(), email(), firstName(), id(), lastName(), status()
  teams:
    primary key: id
    fields: dateCreated(), id(), name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Guru API read of collections, groups, members, and teams
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect guru

  # Inspect as structured JSON
  pm connectors inspect guru --json

AGENT WORKFLOW
  - Run pm connectors inspect guru before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
