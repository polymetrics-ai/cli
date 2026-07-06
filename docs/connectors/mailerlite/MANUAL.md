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
  api_token (secret)

ETL STREAMS
  subscribers:
    primary key: id
    cursor: updated_at
    fields: click_rate(), clicks_count(), created_at(), email(), fields(), id(), ip_address(), open_rate(), opens_count(), sent(), source(), status(), subscribed_at(), unsubscribed_at(), updated_at()
  campaigns:
    primary key: id
    cursor: updated_at
    fields: account_id(), created_at(), finished_at(), id(), is_stopped(), name(), scheduled_for(), started_at(), stats(), status(), type(), updated_at()
  groups:
    primary key: id
    cursor: created_at
    fields: active_count(), click_rate(), clicks_count(), created_at(), id(), name(), open_rate(), opens_count(), sent_count(), unsubscribed_count()
  segments:
    primary key: id
    cursor: created_at
    fields: click_rate(), created_at(), id(), name(), open_rate(), total()
  automations:
    primary key: id
    cursor: created_at
    fields: created_at(), enabled(), id(), name(), stats(), status(), steps(), trigger_data()

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
