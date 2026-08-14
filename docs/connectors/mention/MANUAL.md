# pm connectors inspect mention

```text
NAME
  pm connectors inspect mention - Mention connector manual

SYNOPSIS
  pm connectors inspect mention
  pm connectors inspect mention --json
  pm credentials add <name> --connector mention [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mention app metadata, accounts, alerts, mentions, alert tags, alert shares, alert preferences, and alert tasks from the Mention social listening REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  alert_id
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  app_data:
    fields: actions(object), alert_languages(object), countries(object), days(object), folders(object), integrations(object), languages(object), sources(object), tones(object)
  account_me:
    primary key: id
    fields: created_at(string), id(string), language(string), name(string), permission(string), timezone(string)
  account:
    primary key: id
    fields: created_at(string), id(string), language(string), name(string), permission(string), timezone(string)
  alert:
    primary key: id
    fields: countries(array), created_at(string), description(string), id(string), languages(array), name(string), query(object), sources(array), updated_at(string)
  mention:
    primary key: id
    fields: created_at(string), description(string), favorite(boolean), id(string), language(string), published_at(string), source_name(string), source_type(string), title(string), tone(number), url(string)
  alert_tag:
    primary key: id
    fields: color(string), id(string), name(string)
  alert_share:
    primary key: id
    fields: created_at(string), email(string), id(string), permission(string), updated_at(string)
  alert_preferences:
    fields: frequency(string), notification_frequency(string), send_email(boolean), send_push(boolean), shared(boolean)
  alert_task:
    primary key: id
    fields: created_at(string), description(string), id(string), mention(object), state(string), title(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Mention API read of app metadata, account, alert, mention, tag, share, preference, and task data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mention

  # Inspect as structured JSON
  pm connectors inspect mention --json

AGENT WORKFLOW
  - Run pm connectors inspect mention before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
