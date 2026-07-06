# pm connectors inspect huntr

```text
NAME
  pm connectors inspect huntr - Huntr connector manual

SYNOPSIS
  pm connectors inspect huntr
  pm connectors inspect huntr --json
  pm credentials add <name> --connector huntr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Huntr organization members, candidates, activities, notes, and actions through the Huntr REST API.

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
  page_size
  api_key (secret)

ETL STREAMS
  members:
    primary key: id
    fields: boardIds(), createdAt(), email(), familyName(), fullName(), givenName(), id(), isActive(), lastSeenAt()
  candidates:
    primary key: id
    fields: email(), firstName(), id(), lastName(), memberId()
  activities:
    primary key: id
    fields: activityCategory(), completed(), completedAt(), createdAt(), id(), startAt(), title()
  notes:
    primary key: id
    fields: htmlText(), id(), memberId(), text()
  actions:
    primary key: id
    fields: actionType(), activityId(), candidateId(), createdAt(), date(), id(), memberId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Huntr organization API read of member, candidate, activity, note, and action data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect huntr

  # Inspect as structured JSON
  pm connectors inspect huntr --json

AGENT WORKFLOW
  - Run pm connectors inspect huntr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
