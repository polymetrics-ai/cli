# pm connectors inspect mendeley

```text
NAME
  pm connectors inspect mendeley - Mendeley connector manual

SYNOPSIS
  pm connectors inspect mendeley
  pm connectors inspect mendeley --json
  pm credentials add <name> --connector mendeley [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API.

ICON
  id: simple-icons-mendeley
  asset: icons/simple-icons/mendeley.svg
  title: Mendeley
  simple_icon_slug: mendeley
  simple_icon_hex: 9D1620
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Mendeley
  match: exact-name-or-slug
  matched_by: mendeley

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
  pm connectors inspect mendeley

  # Inspect as structured JSON
  pm connectors inspect mendeley --json

AGENT WORKFLOW
  - Run pm connectors inspect mendeley before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
