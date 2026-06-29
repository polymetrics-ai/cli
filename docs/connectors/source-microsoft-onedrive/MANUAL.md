# pm connectors inspect source-microsoft-onedrive

```text
NAME
  pm connectors inspect source-microsoft-onedrive - Microsoft OneDrive connector manual

SYNOPSIS
  pm connectors inspect source-microsoft-onedrive
  pm connectors inspect source-microsoft-onedrive --json
  pm credentials add <name> --connector source-microsoft-onedrive [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Microsoft OneDrive catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/microsoft-onedrive.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/onedrive/developer/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: file_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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
  OneDrive API reference: https://learn.microsoft.com/en-us/onedrive/developer/
  OneDrive authentication: https://learn.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/authentication
  Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
  Microsoft 365 Status: https://status.office365.com/

CONFIGURATION
  credentials (object) required: Credentials for connecting to the One Drive API
  drive_name (string): Name of the Microsoft OneDrive drive where the file(s) exist.
  folder_path (string): Path to a specific folder within the drives to search for files. Leave empty to search all folders of the drives. This does not apply to shared items.
  search_scope (string): Specifies the location(s) to search for files. Valid options are 'ACCESSIBLE_DRIVES' to search in the selected OneDrive drive, 'SHARED_ITEMS' for shared items the user has acces...
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
  streams (array) required: manual intervention needed
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.tenant_id, credentials.user_principal_name

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-microsoft-onedrive

  # Inspect as JSON
  pm connectors inspect source-microsoft-onedrive --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  OneDrive API reference: https://learn.microsoft.com/en-us/onedrive/developer/
  OneDrive authentication: https://learn.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/authentication
  Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
  Microsoft 365 Status: https://status.office365.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
