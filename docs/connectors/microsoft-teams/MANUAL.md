# pm connectors inspect microsoft-teams

```text
NAME
  pm connectors inspect microsoft-teams - Microsoft Teams connector manual

SYNOPSIS
  pm connectors inspect microsoft-teams
  pm connectors inspect microsoft-teams --json
  pm credentials add <name> --connector microsoft-teams [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Teams users, groups, channels, and device-usage reports through the Microsoft Graph REST API using an OAuth2 client-credentials grant. Read-only.

ICON
  asset: icons/microsoft-teams.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  client_id
  login_base_url
  max_pages
  period
  scope
  token_url
  client_secret (secret)
  tenant_id (secret)

ETL STREAMS
  users:
    primary key: id
    fields: account_enabled(), display_name(), id(), job_title(), mail(), mobile_phone(), office_location(), user_principal_name()
  groups:
    primary key: id
    fields: created_date_time(), description(), display_name(), id(), mail(), mail_enabled(), mail_nickname(), security_enabled(), visibility()
  channels:
    primary key: id
    fields: created_date_time(), description(), display_name(), email(), id(), membership_type(), web_url()
  team_device_usage_report:
    primary key: id
    fields: id(), is_deleted(), last_activity_date(), report_period(), used_android_phone(), used_i_os(), used_mac(), used_web(), used_windows_phone(), user_principal_name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Microsoft Graph API read of tenant users/groups/channels/device-usage data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect microsoft-teams

  # Inspect as structured JSON
  pm connectors inspect microsoft-teams --json

AGENT WORKFLOW
  - Run pm connectors inspect microsoft-teams before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
