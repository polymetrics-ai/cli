# pm connectors inspect mailerlite

```text
NAME
  pm connectors inspect mailerlite - MailerLite connector manual

SYNOPSIS
  pm connectors inspect mailerlite
  pm connectors inspect mailerlite --json
  pm credentials add <name> --connector mailerlite [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads MailerLite subscribers, campaigns, groups, segments, and automations through the MailerLite v2 REST API.

ICON
  id: mailerlite
  asset: icons/mailerlite.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.mailerlite.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_token (secret) (required)

ETL STREAMS
  subscribers:
    primary key: id
    cursor: updated_at
    fields: click_rate(number), clicks_count(integer), created_at(string), email(string), fields(object), id(string), ip_address(string), open_rate(number), opens_count(integer), sent(integer), source(string), status(string), subscribed_at(string), unsubscribed_at(string), updated_at(string)
  campaigns:
    primary key: id
    cursor: updated_at
    fields: account_id(string), created_at(string), finished_at(string), id(string), is_stopped(boolean), name(string), scheduled_for(string), started_at(string), stats(object), status(string), type(string), updated_at(string)
  groups:
    primary key: id
    cursor: created_at
    fields: active_count(integer), click_rate(object), clicks_count(integer), created_at(string), id(string), name(string), open_rate(object), opens_count(integer), sent_count(integer), unsubscribed_count(integer)
  segments:
    primary key: id
    cursor: created_at
    fields: click_rate(object), created_at(string), id(string), name(string), open_rate(object), total(integer)
  automations:
    primary key: id
    cursor: created_at
    fields: created_at(string), enabled(boolean), id(string), name(string), stats(object), status(string), steps(object), trigger_data(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external MailerLite API read of subscriber, campaign, group, segment, and automation data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailerlite

  # Inspect as structured JSON
  pm connectors inspect mailerlite --json

AGENT WORKFLOW
  - Run pm connectors inspect mailerlite before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
