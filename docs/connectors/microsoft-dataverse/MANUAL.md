# pm connectors inspect microsoft-dataverse

```text
NAME
  pm connectors inspect microsoft-dataverse - Microsoft Dataverse connector manual

SYNOPSIS
  pm connectors inspect microsoft-dataverse
  pm connectors inspect microsoft-dataverse --json
  pm credentials add <name> --connector microsoft-dataverse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Dataverse accounts, contacts, leads, opportunities, and users through the Web API.

ICON
  asset: icons/microsoftdataverse.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview

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
  accounts:
    primary key: id
    fields: created_on(), email(), id(), modified_on(), name()
  contacts:
    primary key: id
    fields: created_on(), email(), id(), modified_on(), name()
  leads:
    primary key: id
    fields: created_on(), email(), id(), modified_on(), name()
  opportunities:
    primary key: id
    fields: created_on(), email(), id(), modified_on(), name()
  systemusers:
    primary key: id
    fields: created_on(), email(), id(), modified_on(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Microsoft Dataverse Web API read of CRM records
  approval: none; read-only OAuth2 client-credentials API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect microsoft-dataverse

  # Inspect as structured JSON
  pm connectors inspect microsoft-dataverse --json

AGENT WORKFLOW
  - Run pm connectors inspect microsoft-dataverse before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
