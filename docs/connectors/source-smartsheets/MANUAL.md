# pm connectors inspect source-smartsheets

```text
NAME
  pm connectors inspect source-smartsheets - Smartsheets connector manual

SYNOPSIS
  pm connectors inspect source-smartsheets
  pm connectors inspect source-smartsheets --json
  pm credentials add <name> --connector source-smartsheets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Smartsheets catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/smartsheet.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://smartsheet.redoc.ly/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
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
  family: custom_go_port
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Smartsheet API reference: https://smartsheet.redoc.ly/
  Smartsheet authentication: https://smartsheet.redoc.ly/#section/API-Basics/Authentication-and-Access-Tokens
  Smartsheet rate limits: https://smartsheet.redoc.ly/#section/API-Basics/Rate-Limiting
  Smartsheet Status: https://status.smartsheet.com/

CONFIGURATION
  credentials (object) required
  is_report (boolean): If true, the source will treat the provided sheet_id as a report. If false, the source will treat the provided sheet_id as a sheet.
  metadata_fields (array): A List of available columns which metadata can be pulled from.
  spreadsheet_id (string) required: The spreadsheet ID. Find it by opening the spreadsheet then navigating to File > Properties
  start_datetime (string): Only rows modified after this date/time will be replicated. This should be an ISO 8601 string, for instance: `2000-01-01T13:00:00`
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-smartsheets

  # Inspect as JSON
  pm connectors inspect source-smartsheets --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Smartsheet API reference: https://smartsheet.redoc.ly/
  Smartsheet authentication: https://smartsheet.redoc.ly/#section/API-Basics/Authentication-and-Access-Tokens
  Smartsheet rate limits: https://smartsheet.redoc.ly/#section/API-Basics/Rate-Limiting
  Smartsheet Status: https://status.smartsheet.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
