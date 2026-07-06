# pm connectors inspect bing-ads

```text
NAME
  pm connectors inspect bing-ads - Bing Ads connector manual

SYNOPSIS
  pm connectors inspect bing-ads
  pm connectors inspect bing-ads --json
  pm credentials add <name> --connector bing-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST APIs. Read-only.

ICON
  asset: icons/bingads.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/advertising/guides/release-notes

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
  pm connectors inspect bing-ads

  # Inspect as structured JSON
  pm connectors inspect bing-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect bing-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
