# pm connectors inspect sendinblue

```text
NAME
  pm connectors inspect sendinblue - Sendinblue connector manual

SYNOPSIS
  pm connectors inspect sendinblue
  pm connectors inspect sendinblue --json
  pm credentials add <name> --connector sendinblue [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Sendinblue/Brevo contacts, campaigns, lists, and senders through the Brevo API.

ICON
  asset: icons/sendinblue.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.brevo.com/reference

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: modifiedAt
    fields: email(), id(), modifiedAt()
  email_campaigns:
    primary key: id
    cursor: modifiedAt
    fields: id(), modifiedAt(), name(), status()
  contacts_lists:
    primary key: id
    fields: id(), name()
  senders:
    primary key: id
    fields: email(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Brevo (Sendinblue) API read of contact, campaign, list, and sender data
  approval: none; read-only, no reverse-ETL writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sendinblue

  # Inspect as structured JSON
  pm connectors inspect sendinblue --json

AGENT WORKFLOW
  - Run pm connectors inspect sendinblue before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
