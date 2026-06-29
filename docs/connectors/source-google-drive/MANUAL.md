# pm connectors inspect source-google-drive

```text
NAME
  pm connectors inspect source-google-drive - Google Drive connector manual

SYNOPSIS
  pm connectors inspect source-google-drive
  pm connectors inspect source-google-drive --json
  pm credentials add <name> --connector source-google-drive [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Drive catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/google-drive.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/drive/api/reference/rest/v3

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

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
  Google Drive API reference: https://developers.google.com/drive/api/reference/rest/v3
  Google Drive authentication: https://developers.google.com/drive/api/guides/about-auth
  Google Drive quotas: https://developers.google.com/drive/api/guides/limits
  Google Workspace Status: https://www.google.com/appsstatus/

CONFIGURATION
  credentials (object) required: Credentials for connecting to the Google Drive API
  delivery_method (object)
  folder_url (string) required: URL for the folder you want to sync. Using individual streams and glob patterns, it's possible to only sync a subset of all files located in the folder.
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
  streams (array) required: manual intervention needed
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account_info

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-google-drive

  # Inspect as JSON
  pm connectors inspect source-google-drive --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Drive API reference: https://developers.google.com/drive/api/reference/rest/v3
  Google Drive authentication: https://developers.google.com/drive/api/guides/about-auth
  Google Drive quotas: https://developers.google.com/drive/api/guides/limits
  Google Workspace Status: https://www.google.com/appsstatus/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
