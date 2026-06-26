# pm connectors inspect web-scrapper

```text
NAME
  pm connectors inspect web-scrapper - Web Scrapper connector manual

SYNOPSIS
  pm connectors inspect web-scrapper
  pm connectors inspect web-scrapper --json
  pm credentials add <name> --connector web-scrapper [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads sitemap and scraping job metadata from the Web Scraper Cloud API.

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
  pm connectors inspect web-scrapper

  # Inspect as structured JSON
  pm connectors inspect web-scrapper --json

AGENT WORKFLOW
  - Run pm connectors inspect web-scrapper before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
