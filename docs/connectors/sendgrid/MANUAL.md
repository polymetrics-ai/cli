# pm connectors inspect sendgrid

```text
NAME
  pm connectors inspect sendgrid - Sendgrid connector manual

SYNOPSIS
  pm connectors inspect sendgrid
  pm connectors inspect sendgrid --json
  pm credentials add <name> --connector sendgrid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SendGrid Marketing Campaigns lists, segments, and contacts, plus suppression bounces, through the SendGrid v3 REST API. Read-only.

ICON
  id: sendgrid
  asset: icons/sendgrid.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  lists:
    primary key: id
    fields: contact_count(integer), id(string), name(string)
  segments:
    primary key: id
    cursor: updated_at
    fields: contacts_count(integer), created_at(string), id(string), name(string), query_version(string), sample_updated_at(string), updated_at(string)
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), first_name(string), id(string), last_name(string), phone_number(string), updated_at(string)
  suppression_bounces:
    primary key: email
    cursor: created
    fields: created(integer), email(string), reason(string), status(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external SendGrid API read of marketing list, segment, contact, and suppression-bounce data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sendgrid

  # Inspect as structured JSON
  pm connectors inspect sendgrid --json

AGENT WORKFLOW
  - Run pm connectors inspect sendgrid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
