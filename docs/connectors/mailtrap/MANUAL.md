# pm connectors inspect mailtrap

```text
NAME
  pm connectors inspect mailtrap - Mailtrap connector manual

SYNOPSIS
  pm connectors inspect mailtrap
  pm connectors inspect mailtrap --json
  pm credentials add <name> --connector mailtrap [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap account-management REST API.

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
  account_id
  base_url
  api_token (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: access_levels(), id(), name()
  inboxes:
    primary key: id
    fields: account_id(), domain(), email_username(), emails_count(), id(), max_size(), name(), status(), used_size()
  projects:
    primary key: id
    fields: account_id(), id(), name()
  sending_domains:
    primary key: id
    fields: account_id(), demo(), domain_name(), id(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mailtrap API read of account-management data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailtrap

  # Inspect as structured JSON
  pm connectors inspect mailtrap --json

AGENT WORKFLOW
  - Run pm connectors inspect mailtrap before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
