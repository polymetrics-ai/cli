# pm connectors inspect google-calendar

```text
NAME
  pm connectors inspect google-calendar - Google Calendar connector manual

SYNOPSIS
  pm connectors inspect google-calendar
  pm connectors inspect google-calendar --json
  pm credentials add <name> --connector google-calendar [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Calendar calendar lists, events, settings, and access control rules through the Calendar API v3 using an OAuth2 refresh token.

ICON
  id: simple-icons-googlecalendar
  asset: icons/simple-icons/googlecalendar.svg
  title: Google Calendar
  simple_icon_slug: googlecalendar
  simple_icon_hex: 4285F4
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Calendar
  match: exact-name-or-slug
  matched_by: google-calendar

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
  pm connectors inspect google-calendar

  # Inspect as structured JSON
  pm connectors inspect google-calendar --json

AGENT WORKFLOW
  - Run pm connectors inspect google-calendar before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
