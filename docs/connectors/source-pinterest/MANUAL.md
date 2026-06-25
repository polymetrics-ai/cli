# pm connectors inspect source-pinterest

```text
NAME
  pm connectors inspect source-pinterest - Pinterest connector manual

SYNOPSIS
  pm connectors inspect source-pinterest
  pm connectors inspect source-pinterest --json
  pm credentials add <name> --connector source-pinterest [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pinterest catalog connector for https://docs.airbyte.com/integrations/sources/pinterest. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-pinterest:2.2.4 (metadata only; not executed)

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
  family: declarative_http_source
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Changelog: https://developers.pinterest.com/docs/changelog/changelog/
  Pinterest API Changelog: https://developers.pinterest.com/docs/changelog/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/pinterest

CONFIGURATION
  account_id (string): The Pinterest account ID you want to fetch data for. This ID must be provided to filter the data for a specific account.
  credentials (object)
  custom_reports (array): A list which contains ad statistics entries, each entry must have a name and can contains fields, breakdowns or action_breakdowns. Click on "add" to fill this field.
  num_workers (integer): The number of worker threads to use for the sync. Higher values can speed up syncs but may increase rate-limit pressure against Pinterest.
  start_date (string): A date in the format YYYY-MM-DD. If you have not set a date, it would be defaulted to latest allowed date by api (89 days from today).
  status ([array null]): For the ads, ad_groups, and campaigns streams, specifying a status will filter out records that do not match the specified ones. If a status is not specified, the source will de...
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/pinterest

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pinterest

  # Inspect as JSON
  pm connectors inspect source-pinterest --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Pinterest documentation: https://docs.airbyte.com/integrations/sources/pinterest

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
