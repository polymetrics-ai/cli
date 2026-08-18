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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  peoples:
    primary key: id
    fields: company(string), country(string), createdAt(string), email(string), firstName(string), id(string), lastName(string), name(string), phone(string), title(string), updatedAt(string), userId(string)
  segments:
    primary key: id
    fields: createdAt(string), id(string), name(string), type(string), updatedAt(string)
  fields:
    primary key: name
    fields: format(string), name(string), title(string), type(string)
  account_tags:
    primary key: tag
    fields: createdAt(string), id(string), tag(string)
  schemas:
    primary key: name
    fields: name(string), title(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
