# pm connectors inspect gorgias

```text
NAME
  pm connectors inspect gorgias - Gorgias connector manual

SYNOPSIS
  pm connectors inspect gorgias
  pm connectors inspect gorgias --json
  pm credentials add <name> --connector gorgias [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias REST API (read-only).

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
  page_size
  username
  password (secret)

ETL STREAMS
  tickets:
    primary key: id
    cursor: updated_datetime
    fields: channel(), closed_datetime(), created_datetime(), id(), is_unread(), language(), opened_datetime(), priority(), spam(), status(), subject(), trashed_datetime(), updated_datetime(), via()
  customers:
    primary key: id
    cursor: updated_datetime
    fields: channel(), created_datetime(), email(), external_id(), firstname(), id(), language(), lastname(), name(), timezone(), updated_datetime()
  messages:
    primary key: id
    cursor: created_datetime
    fields: body_text(), channel(), created_datetime(), from_agent(), id(), public(), sent_datetime(), stripped_text(), subject(), ticket_id(), via()
  satisfaction_surveys:
    primary key: id
    cursor: created_datetime
    fields: body_text(), created_datetime(), customer_id(), id(), scale_range(), score(), scored_datetime(), sent_datetime(), ticket_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Gorgias API read of helpdesk tickets, customers, messages, and satisfaction surveys
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gorgias

  # Inspect as structured JSON
  pm connectors inspect gorgias --json

AGENT WORKFLOW
  - Run pm connectors inspect gorgias before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
