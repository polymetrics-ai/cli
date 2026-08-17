# pm connectors inspect hibob

```text
NAME
  pm connectors inspect hibob - HiBob connector manual

SYNOPSIS
  pm connectors inspect hibob
  pm connectors inspect hibob --json
  pm credentials add <name> --connector hibob [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads HiBob HR data: employee profiles, company named lists, and people field definitions via the HiBob REST API (read-only).

ICON
  asset: icons/hibob.svg
  source: official
  review_status: official_verified
  review_url: https://apidocs.hibob.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  username
  password (secret)

ETL STREAMS
  profiles:
    primary key: id
    fields: displayName(), email(), firstName(), fullName(), id(), personal_pronouns(), surname(), work_department(), work_isManager(), work_site(), work_startDate(), work_title()
  named_lists:
    primary key: id
    fields: archived(), children(), id(), name(), parentId(), value()
  company_lists:
    primary key: id
    fields: category(), description(), id(), name(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external HiBob API read of employee profile and HR metadata
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hibob

  # Inspect as structured JSON
  pm connectors inspect hibob --json

AGENT WORKFLOW
  - Run pm connectors inspect hibob before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
