# pm connectors inspect destination-google-sheets

```text
NAME
  pm connectors inspect destination-google-sheets - Google Sheets connector manual

SYNOPSIS
  pm connectors inspect destination-google-sheets
  pm connectors inspect destination-google-sheets --json
  pm credentials add <name> --connector destination-google-sheets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Sheets catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/google-sheets.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
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
  family: destination_writer
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  Google Sheets documentation: https://developers.google.com/workspace/sheets/api/guides/concepts

CONFIGURATION
  credentials (object) required: Authentication method to access Google Sheets
  spreadsheet_id (string) required: manual intervention needed
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account_info

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-google-sheets

  # Inspect as JSON
  pm connectors inspect destination-google-sheets --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Sheets documentation: https://developers.google.com/workspace/sheets/api/guides/concepts

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
