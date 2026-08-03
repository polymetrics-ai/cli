# pm connectors inspect google-classroom

```text
NAME
  pm connectors inspect google-classroom - Google Classroom connector manual

SYNOPSIS
  pm connectors inspect google-classroom
  pm connectors inspect google-classroom --json
  pm credentials add <name> --connector google-classroom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Classroom courses, teachers, students, course work, and announcements through the Classroom REST API using an OAuth2 refresh token.

ICON
  id: simple-icons-googleclassroom
  asset: icons/simple-icons/googleclassroom.svg
  title: Google Classroom
  simple_icon_slug: googleclassroom
  simple_icon_hex: 0F9D58
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Classroom
  match: exact-name-or-slug
  matched_by: google-classroom

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
  pm connectors inspect google-classroom

  # Inspect as structured JSON
  pm connectors inspect google-classroom --json

AGENT WORKFLOW
  - Run pm connectors inspect google-classroom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
