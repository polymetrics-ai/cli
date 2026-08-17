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
  asset: icons/salesforce.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/rest_rns.htm

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_version
  instance_url
  mode
  access_token (secret)

ETL STREAMS
  sobjects:
    primary key: qualified_api_name
    fields: label(), qualified_api_name()
  accounts:
    primary key: id
    fields: email(), id(), last_modified_date(), name()
  contacts:
    primary key: id
    fields: email(), id(), last_modified_date(), name()
  leads:
    primary key: id
    fields: email(), id(), last_modified_date(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
