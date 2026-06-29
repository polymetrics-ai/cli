# pm connectors inspect source-google-sheets

```text
NAME
  pm connectors inspect source-google-sheets - Google Sheets connector manual

SYNOPSIS
  pm connectors inspect source-google-sheets
  pm connectors inspect source-google-sheets --json
  pm credentials add <name> --connector source-google-sheets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Sheets catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/google-sheets.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/workspace/release-notes

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Google Workspace developer release notes: https://developers.google.com/workspace/release-notes
  Release notes: https://developers.google.com/sheets/docs/release-notes

CONFIGURATION
  allow_leading_numbers (boolean): Allows column names to start with numbers. Example: "50th Percentile" → "50_th_percentile" This option will only work if "Convert Column Names to SQL-Compliant Format (names_c...
  batch_size (integer): Default value is 1000000. An integer representing row batch size for each sent request to Google Sheets API. Row batch size means how many rows are processed from the google she...
  combine_letter_number_pairs (boolean): Combines adjacent letters and numbers. Example: "Q3 2023" → "q3_2023" This option will only work if "Convert Column Names to SQL-Compliant Format (names_conversion)" is enabled.
  combine_number_word_pairs (boolean): Combines adjacent numbers and words. Example: "50th Percentile?" → "_50th_percentile_" This option will only work if "Convert Column Names to SQL-Compliant Format (names_conve...
  credentials (object) required: Credentials for connecting to the Google Sheets API
  names_conversion (boolean): Converts column names to a SQL-compliant format (snake_case, lowercase, etc). If enabled, you can further customize the sanitization using the options below.
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs for spreadsheets with multiple sheets, but may hit rate limits. Google Sheets API limits to 300 read r...
  read_empty_header_columns (boolean): When enabled, the connector will continue reading columns after empty header cells and will include data from those columns using generated column names (e.g., "column_C"). By d...
  remove_leading_trailing_underscores (boolean): Removes leading and trailing underscores from column names. Does not remove leading underscores from column names that start with a number. Example: "50th Percentile? "→ "_50_...
  remove_special_characters (boolean): Removes all special characters from column names. Example: "Example ID*" → "example_id" This option will only work if "Convert Column Names to SQL-Compliant Format (names_conv...
  spreadsheet_id (string) required: Enter the link to the Google spreadsheet you want to sync. To copy the link, click the 'Share' button in the top-right corner of the spreadsheet, then click 'Copy link'.
  stream_name_overrides (array): manual intervention needed
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
  pm connectors inspect source-google-sheets

  # Inspect as JSON
  pm connectors inspect source-google-sheets --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Workspace developer release notes: https://developers.google.com/workspace/release-notes
  Release notes: https://developers.google.com/sheets/docs/release-notes

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
