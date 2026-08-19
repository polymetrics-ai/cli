# pm connectors inspect salesforce

```text
NAME
  pm connectors inspect salesforce - Salesforce connector manual

SYNOPSIS
  pm connectors inspect salesforce
  pm connectors inspect salesforce --json
  pm credentials add <name> --connector salesforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Salesforce object metadata and allow-listed Account, Contact, and Lead SOQL queries through the REST API. Read-only.

ICON
  id: salesforce
  asset: icons/salesforce.svg
  source: official
  review_status: official_verified
  review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_version
  instance_url (required)
  mode
  access_token (secret) (required)

ETL STREAMS
  sobjects:
    primary key: qualified_api_name
    fields: label(string), qualified_api_name(string)
  accounts:
    primary key: id
    fields: email(string), id(string), last_modified_date(string), name(string)
  contacts:
    primary key: id
    fields: email(string), id(string), last_modified_date(string), name(string)
  leads:
    primary key: id
    fields: email(string), id(string), last_modified_date(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Salesforce API read of object metadata, Account, Contact, and Lead records
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect salesforce

  # Inspect as structured JSON
  pm connectors inspect salesforce --json

AGENT WORKFLOW
  - Run pm connectors inspect salesforce before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
