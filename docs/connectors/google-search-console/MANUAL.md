# pm connectors inspect google-search-console

```text
NAME
  pm connectors inspect google-search-console - Google Search Console connector manual

SYNOPSIS
  pm connectors inspect google-search-console
  pm connectors inspect google-search-console --json
  pm credentials add <name> --connector google-search-console [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date, query, page, country, and device) through the Search Console v3 REST API. Read-only.

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
  pm connectors inspect google-search-console

  # Inspect as structured JSON
  pm connectors inspect google-search-console --json

AGENT WORKFLOW
  - Run pm connectors inspect google-search-console before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
