# pm connectors inspect lever-hiring

```text
NAME
  pm connectors inspect lever-hiring - Lever Hiring connector manual

SYNOPSIS
  pm connectors inspect lever-hiring
  pm connectors inspect lever-hiring --json
  pm credentials add <name> --connector lever-hiring [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data API. Read-only (full-refresh).

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
  access_token (secret)
  api_key (secret)

ETL STREAMS
  opportunities:
    primary key: id
    cursor: createdAt
    fields: archivedAt(), createdAt(), emails(), headline(), id(), lastInteractionAt(), name(), origin(), sources(), stage(), tags(), updatedAt()
  postings:
    primary key: id
    cursor: createdAt
    fields: categories(), createdAt(), hiringManager(), id(), owner(), state(), text(), updatedAt(), user()
  users:
    primary key: id
    cursor: createdAt
    fields: accessRole(), createdAt(), deactivatedAt(), email(), id(), name(), username()
  requisitions:
    primary key: id
    cursor: createdAt
    fields: createdAt(), headcountHired(), headcountTotal(), id(), name(), owner(), requisitionCode(), status(), updatedAt()
  stages:
    primary key: id
    fields: id(), text()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Lever API read of candidate and hiring pipeline data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect lever-hiring

  # Inspect as structured JSON
  pm connectors inspect lever-hiring --json

AGENT WORKFLOW
  - Run pm connectors inspect lever-hiring before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
