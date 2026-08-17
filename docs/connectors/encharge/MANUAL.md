# pm connectors inspect encharge

```text
NAME
  pm connectors inspect encharge - Encharge connector manual

SYNOPSIS
  pm connectors inspect encharge
  pm connectors inspect encharge --json
  pm credentials add <name> --connector encharge [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Encharge people, segments, fields, account tags, and schemas through the Encharge REST API.

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
  api_key (secret)

ETL STREAMS
  peoples:
    primary key: id
    fields: company(), country(), createdAt(), email(), firstName(), id(), lastName(), name(), phone(), title(), updatedAt(), userId()
  segments:
    primary key: id
    fields: createdAt(), id(), name(), type(), updatedAt()
  fields:
    primary key: name
    fields: format(), name(), title(), type()
  account_tags:
    primary key: tag
    fields: createdAt(), id(), tag()
  schemas:
    primary key: name
    fields: name(), title(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Encharge API read of people, segment, field, and tag data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect encharge

  # Inspect as structured JSON
  pm connectors inspect encharge --json

AGENT WORKFLOW
  - Run pm connectors inspect encharge before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
