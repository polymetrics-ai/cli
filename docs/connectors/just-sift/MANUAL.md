# pm connectors inspect just-sift

```text
NAME
  pm connectors inspect just-sift - JustSift connector manual

SYNOPSIS
  pm connectors inspect just-sift
  pm connectors inspect just-sift --json
  pm credentials add <name> --connector just-sift [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads JustSift people directory profiles and person field definitions through the Sift REST API.

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
  mode
  api_token (secret)

ETL STREAMS
  peoples:
    primary key: id
    fields: companyName(), connector(), department(), directReportCount(), directoryId(), displayName(), email(), firstName(), id(), isTeamLeader(), lastName(), officeCity(), officeState(), phone(), pictureUrl(), title()
  fields:
    primary key: id
    fields: connector(), displayName(), filterable(), id(), objectKey(), searchable(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external JustSift API read of people directory profiles and field definitions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect just-sift

  # Inspect as structured JSON
  pm connectors inspect just-sift --json

AGENT WORKFLOW
  - Run pm connectors inspect just-sift before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
