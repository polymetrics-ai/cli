# pm connectors inspect source-sharepoint-enterprise

```text
NAME
  pm connectors inspect source-sharepoint-enterprise - SharePoint Enterprise connector manual

SYNOPSIS
  pm connectors inspect source-sharepoint-enterprise
  pm connectors inspect source-sharepoint-enterprise --json
  pm credentials add <name> --connector source-sharepoint-enterprise [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SharePoint Enterprise catalog connector for https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-enterprise. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: file_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-sharepoint-enterprise:0.3.2 (metadata only; not executed)

RUNTIME CAPABILITIES
  metadata=true
  check=false
  catalog=false
  read=false
  write=false
  query=false
  etl=false
  reverse_etl=false
  unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

NATIVE PORT PLAN
  family: file_object_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-enterprise

CONFIGURATION
  credentials (object) required: Credentials for connecting to the One Drive API
  delivery_method (object)
  file_contains_query (array): Input additional query to search files. It will make search files step faster if your Sharepoint account has a lot of files and folders. This query text will be used in the requ...
  folder_path (string): Path to a specific folder within the drives to search for files. Leave empty to search all folders of the drives. This does not apply to shared items.
  search_scope (string): Specifies the location(s) to search for files. Valid options are 'ACCESSIBLE_DRIVES' for all SharePoint drives the user can access, 'SHARED_ITEMS' for shared items the user has ...
  site_url (string): Url of SharePoint site to search for files. Leave empty to search in the main site. Use 'https://<tenant_name>.sharepoint.com/sites/' to iterate over all sites.
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
  streams (array) required: Each instance of this configuration defines a <a href="https://docs.airbyte.com/cloud/core-concepts#stream">stream</a>. Use this to define which files belong in the stream, thei...
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.tenant_id, credentials.user_principal_name

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-enterprise

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sharepoint-enterprise

  # Inspect as JSON
  pm connectors inspect source-sharepoint-enterprise --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SharePoint Enterprise documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-enterprise

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
