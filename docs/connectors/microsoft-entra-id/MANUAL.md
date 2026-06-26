# pm connectors inspect microsoft-entra-id

```text
NAME
  pm connectors inspect microsoft-entra-id - Microsoft Entra ID connector manual

SYNOPSIS
  pm connectors inspect microsoft-entra-id
  pm connectors inspect microsoft-entra-id --json
  pm credentials add <name> --connector microsoft-entra-id [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Entra ID (Azure AD) directory objects — users, groups, applications, service principals, and directory roles — from the Microsoft Graph API using OAuth2 client credentials.

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

SECURITY
  read risk: connector-specific
  write risk: connector-specific
  approval: external mutations require preview and approval
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
