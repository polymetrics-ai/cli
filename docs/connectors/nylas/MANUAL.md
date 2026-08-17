# pm connectors inspect nylas

```text
NAME
  pm connectors inspect nylas - Nylas connector manual

SYNOPSIS
  pm connectors inspect nylas
  pm connectors inspect nylas --json
  pm credentials add <name> --connector nylas [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Nylas calendars, contacts, messages, and events for a connected grant through the Nylas v3 REST API.

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
  calendar_id
  grant_id
  max_pages
  mode
  page_size
  api_key (secret)

ETL STREAMS
  calendars:
    primary key: id
    fields: description(), grant_id(), hex_color(), id(), is_primary(), name(), object(), read_only(), timezone()
  contacts:
    primary key: id
    fields: company_name(), emails(), given_name(), grant_id(), id(), job_title(), object(), phone_numbers(), source(), surname()
  messages:
    primary key: id
    cursor: date
    fields: date(), folders(), from(), grant_id(), id(), object(), snippet(), starred(), subject(), thread_id(), to(), unread()
  events:
    primary key: id
    cursor: updated_at
    fields: busy(), calendar_id(), description(), grant_id(), id(), location(), object(), read_only(), status(), title(), updated_at(), when()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Nylas API read of a connected grant's calendar, contact, and message data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nylas

  # Inspect as structured JSON
  pm connectors inspect nylas --json

AGENT WORKFLOW
  - Run pm connectors inspect nylas before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
