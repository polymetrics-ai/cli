# pm connectors inspect marketo

```text
NAME
  pm connectors inspect marketo - Marketo connector manual

SYNOPSIS
  pm connectors inspect marketo
  pm connectors inspect marketo --json
  pm credentials add <name> --connector marketo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Marketo leads, programs, and activities through Marketo REST endpoints. Read-only; does not refresh OAuth tokens internally.

ICON
  asset: icons/marketo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.marketo.com/rest-api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  activity_type_ids
  base_url
  max_pages
  mode
  page_size
  access_token (secret)

ETL STREAMS
  leads:
    primary key: id
    fields: createdAt(), email(), id(), updatedAt()
  programs:
    primary key: id
    fields: createdAt(), id(), name(), updatedAt()
  activities:
    primary key: id
    fields: activityDate(), activityTypeId(), id(), leadId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Marketo REST API read of lead, program, and activity data
  approval: none; read-only Marketo REST API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect marketo

  # Inspect as structured JSON
  pm connectors inspect marketo --json

AGENT WORKFLOW
  - Run pm connectors inspect marketo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
