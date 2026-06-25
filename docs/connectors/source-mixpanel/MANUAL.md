# pm connectors inspect source-mixpanel

```text
NAME
  pm connectors inspect source-mixpanel - Mixpanel connector manual

SYNOPSIS
  pm connectors inspect source-mixpanel
  pm connectors inspect source-mixpanel --json
  pm credentials add <name> --connector source-mixpanel [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Mixpanel catalog connector for https://docs.airbyte.com/integrations/sources/mixpanel. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-mixpanel:4.0.0 (metadata only; not executed)

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Mixpanel API reference: https://developer.mixpanel.com/reference/overview
  Mixpanel authentication: https://developer.mixpanel.com/reference/authentication
  Mixpanel rate limits: https://developer.mixpanel.com/reference/rate-limits
  Mixpanel Status: https://status.mixpanel.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/mixpanel

CONFIGURATION
  attribution_window (integer): A period of time for attributing results to ads and the lookback period after those actions occur during which ad results are counted. Default attribution window is 5 days. (Thi...
  credentials (object) required: Choose how to authenticate to Mixpanel
  date_window_size (integer): Defines window size in days, that used to slice through data. You can reduce it, if amount of data in each window is too big for your environment. (This value should be positive...
  end_date (string): The date in the format YYYY-MM-DD. Any data after this date will not be replicated. Left empty to always sync to most recent date
  export_lookback_window (integer): The number of seconds to look back from the last synced timestamp during incremental syncs of the Export stream. This ensures no data is missed due to delays in event recording....
  num_workers (integer): The number of worker threads to use for the sync. The performance upper boundary is based on the limit of your Mixpanel pricing plan. More info about the rate limit tiers can be...
  page_size (integer): The number of records to fetch per request for the engage stream. Default is 1000. If you are experiencing long sync times with this stream, try increasing this value.
  project_timezone (string): Time zone in which integer date times are stored. The project timezone may be found in the project settings in the <a href="https://help.mixpanel.com/hc/en-us/articles/115004547...
  region (string): The region of mixpanel domain instance either US or EU.
  select_properties_by_default (boolean): Setting this config parameter to TRUE ensures that new properties on events and engage records are captured. Otherwise new properties will be ignored.
  start_date (string): The date in the format YYYY-MM-DD. Any data before this date will not be replicated. If this option is not set, the connector will replicate data from up to one year ago by defa...
  secret fields: credentials.api_secret, credentials.secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/mixpanel

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-mixpanel

  # Inspect as JSON
  pm connectors inspect source-mixpanel --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Mixpanel documentation: https://docs.airbyte.com/integrations/sources/mixpanel

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
