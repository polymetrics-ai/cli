# pm connectors inspect intruder

```text
NAME
  pm connectors inspect intruder - Intruder connector manual

SYNOPSIS
  pm connectors inspect intruder
  pm connectors inspect intruder --json
  pm credentials add <name> --connector intruder [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Intruder issues, issue occurrences, scans, and targets through the Intruder REST API (read-only, full refresh).

ICON
  asset: icons/intruder.svg
  source: official
  review_status: official_verified
  review_url: https://developers.intruder.io/docs/welcome

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
  access_token (secret)

ETL STREAMS
  issues:
    primary key: id
    fields: description(), id(), occurrences(), remediation(), severity(), snooze_reason(), snooze_until(), snoozed(), title()
  scans:
    primary key: id
    fields: created_at(), id(), status()
  targets:
    primary key: id
    fields: address(), id(), tags()
  occurrences:
    primary key: id
    fields: age(), extra_info(), id(), issue_id(), port(), snooze_reason(), snooze_until(), snoozed(), target()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Intruder API read of vulnerability issues, issue occurrences, scans, and target data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect intruder

  # Inspect as structured JSON
  pm connectors inspect intruder --json

AGENT WORKFLOW
  - Run pm connectors inspect intruder before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
