# pm connectors inspect source-outbrain-amplify

```text
NAME
  pm connectors inspect source-outbrain-amplify - Outbrain Amplify connector manual

SYNOPSIS
  pm connectors inspect source-outbrain-amplify
  pm connectors inspect source-outbrain-amplify --json
  pm credentials add <name> --connector source-outbrain-amplify [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Outbrain Amplify catalog connector for https://docs.airbyte.com/integrations/sources/outbrain-amplify. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-outbrain-amplify:0.2.21 (metadata only; not executed)

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
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/outbrain-amplify

CONFIGURATION
  conversion_count (string): The definition of conversion count in reports. See <a href="https://amplifyv01.docs.apiary.io/#reference/performance-reporting/periodic/retrieve-performance-statistics-for-all-m...
  credentials (object) required: Credentials for making authenticated requests requires either username/password or access_token.
  end_date (string): Date in the format YYYY-MM-DD.
  geo_location_breakdown (string): The granularity used for geo location data in reports.
  report_granularity (string): The granularity used for periodic data in reports. See <a href="https://amplifyv01.docs.apiary.io/#reference/performance-reporting/periodic/retrieve-performance-statistics-for-a...
  start_date (string) required: Date in the format YYYY-MM-DD eg. 2017-01-25. Any data before this date will not be replicated.
  secret fields: credentials.access_token, credentials.password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/outbrain-amplify

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-outbrain-amplify

  # Inspect as JSON
  pm connectors inspect source-outbrain-amplify --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Outbrain Amplify documentation: https://docs.airbyte.com/integrations/sources/outbrain-amplify

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
