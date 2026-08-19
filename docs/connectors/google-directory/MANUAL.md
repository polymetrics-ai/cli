# pm connectors inspect google-directory

```text
NAME
  pm connectors inspect google-directory - Google Directory connector manual

SYNOPSIS
  pm connectors inspect google-directory
  pm connectors inspect google-directory --json
  pm credentials add <name> --connector google-directory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Admin SDK Directory users, groups, organizational units, and ChromeOS devices via bearer-token OAuth. Read-only.

ICON
  id: googledirectory
  asset: icons/googledirectory.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/admin-sdk/directory/reference/rest

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  customer_id
  max_pages
  mode
  page_size
  access_token (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: id(string), name(string), org_unit_path(string), primary_email(string)
  groups:
    primary key: id
    fields: description(string), email(string), id(string), name(string)
  orgunits:
    primary key: id
    fields: description(string), id(string), name(string), org_unit_path(string)
  chromeos_devices:
    primary key: id
    fields: id(string), org_unit_path(string), serial_number(string), status(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Google Admin SDK Directory API read of user/group/org-unit/device metadata
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-directory

  # Inspect as structured JSON
  pm connectors inspect google-directory --json

AGENT WORKFLOW
  - Run pm connectors inspect google-directory before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
