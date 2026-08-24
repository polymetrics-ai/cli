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

COMMAND SURFACE
  Declared salesforce API commands.
  Usage: pm salesforce <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Other Commands
    operations get-services-data-version-sobjects - Declared etl: GET /services/data/{version}/sobjects. [intent=etl availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-sobjects-1 has no declaration-owned executable etl route.
    operations get-services-data-version-query-select-from-account - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Account). [intent=etl availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-query-select-from-account-2 has no declaration-owned executable etl route.
    operations get-services-data-version-query-select-from-contact - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Contact). [intent=etl availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-query-select-from-contact-3 has no declaration-owned executable etl route.
    operations get-services-data-version-query-select-from-lead - Declared etl: GET /services/data/{version}/query (SELECT ... FROM Lead). [intent=etl availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-query-select-from-lead-4 has no declaration-owned executable etl route.
    operations get-services-data-version-query-arbitrary-soql - Declared direct read: GET /services/data/{version}/query (arbitrary SOQL). [intent=direct_read availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-query-arbitrary-soql-5 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations get-services-data-version-sobjects-s-object-describe - Declared direct read: GET /services/data/{version}/sobjects/{sObject}/describe. [intent=direct_read availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-sobjects-sobject-describe-6 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
    operations post-services-data-version-sobjects-s-object - Declared direct write: POST /services/data/{version}/sobjects/{sObject}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /services/data/{version}/sobjects/{sObject}.; notes: Blocked: locked source operation salesforce.local-api-surface.post-services-data-version-sobjects-sobject-7 has no declaration-owned executable direct_write route.
    operations patch-services-data-version-sobjects-s-object-id - Declared direct write: PATCH /services/data/{version}/sobjects/{sObject}/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: PATCH /services/data/{version}/sobjects/{sObject}/{id}.; notes: Blocked: locked source operation salesforce.local-api-surface.patch-services-data-version-sobjects-sobject-id-8 has no declaration-owned executable direct_write route.
    operations delete-services-data-version-sobjects-s-object-id - Declared direct write: DELETE /services/data/{version}/sobjects/{sObject}/{id}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /services/data/{version}/sobjects/{sObject}/{id}.; notes: Blocked: locked source operation salesforce.local-api-surface.delete-services-data-version-sobjects-sobject-id-9 has no declaration-owned executable direct_write route.
    operations get-services-data-version-jobs-query-bulk-api - Declared direct read: GET /services/data/{version}/jobs/query (Bulk API). [intent=direct_read availability=partial]; notes: Blocked: locked source operation salesforce.local-api-surface.get-services-data-version-jobs-query-bulk-api-10 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor

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
