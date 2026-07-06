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
  alert_id
  base_url
  mode
  api_key (secret)

ETL STREAMS
  app_data:
    fields: actions(), alert_languages(), countries(), days(), folders(), integrations(), languages(), sources(), tones()
  account_me:
    primary key: id
    fields: created_at(), id(), language(), name(), permission(), timezone()
  account:
    primary key: id
    fields: created_at(), id(), language(), name(), permission(), timezone()
  alert:
    primary key: id
    fields: countries(), created_at(), description(), id(), languages(), name(), query(), sources(), updated_at()
  mention:
    primary key: id
    fields: created_at(), description(), favorite(), id(), language(), published_at(), source_name(), source_type(), title(), tone(), url()
  alert_tag:
    primary key: id
    fields: color(), id(), name()
  alert_share:
    primary key: id
    fields: created_at(), email(), id(), permission(), updated_at()
  alert_preferences:
    fields: frequency(), notification_frequency(), send_email(), send_push(), shared()
  alert_task:
    primary key: id
    fields: created_at(), description(), id(), mention(), state(), title(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
