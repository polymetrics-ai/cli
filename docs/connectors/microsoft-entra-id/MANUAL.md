# pm connectors inspect microsoft-entra-id

```text
NAME
  pm connectors inspect microsoft-entra-id - Microsoft Entra ID connector manual

SYNOPSIS
  pm connectors inspect microsoft-entra-id
  pm connectors inspect microsoft-entra-id --json
  pm credentials add <name> --connector microsoft-entra-id [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Entra ID (Azure AD) directory objects — users, groups, applications, service principals, and directory roles — from the Microsoft Graph API using an OAuth2 client-credentials grant. Read-only.

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
  login_base_url
  max_pages
  mode
  page_size
  scope
  token_url
  client_id (secret)
  client_secret (secret)
  tenant_id (secret)

ETL STREAMS
  users:
    primary key: id
    fields: account_enabled(), department(), display_name(), given_name(), id(), job_title(), mail(), mobile_phone(), office_location(), surname(), user_principal_name()
  groups:
    primary key: id
    fields: created_date_time(), description(), display_name(), id(), mail(), mail_enabled(), mail_nickname(), security_enabled(), visibility()
  applications:
    primary key: id
    fields: app_id(), created_date_time(), description(), display_name(), id(), publisher_domain(), sign_in_audience()
  serviceprincipals:
    primary key: id
    fields: account_enabled(), app_id(), app_owner_organization_id(), display_name(), id(), service_principal_type(), sign_in_audience()
  directoryroles:
    primary key: id
    fields: description(), display_name(), id(), role_template_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Microsoft Graph API read of tenant directory (users/groups/applications/service principals/directory roles) data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect microsoft-entra-id

  # Inspect as structured JSON
  pm connectors inspect microsoft-entra-id --json

AGENT WORKFLOW
  - Run pm connectors inspect microsoft-entra-id before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
