# pm connectors inspect criteo-marketing

```text
NAME
  pm connectors inspect criteo-marketing - Criteo Marketing connector manual

SYNOPSIS
  pm connectors inspect criteo-marketing
  pm connectors inspect criteo-marketing --json
  pm credentials add <name> --connector criteo-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Criteo Marketing Solutions ad sets, advertisers, campaigns, audiences, and ad spend statistics through the Criteo REST API using OAuth2 client-credentials auth.

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
  pm connectors inspect criteo-marketing

  # Inspect as structured JSON
  pm connectors inspect criteo-marketing --json

AGENT WORKFLOW
  - Run pm connectors inspect criteo-marketing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
