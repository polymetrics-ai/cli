# pm connectors inspect source-pagerduty

```text
NAME
  pm connectors inspect source-pagerduty - Pagerduty connector manual

SYNOPSIS
  pm connectors inspect source-pagerduty
  pm connectors inspect source-pagerduty --json
  pm credentials add <name> --connector source-pagerduty [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pagerduty catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pagerduty.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.pagerduty.com/api-reference/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
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
  family: declarative_http_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  PagerDuty API reference: https://developer.pagerduty.com/api-reference/
  PagerDuty authentication: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUw-authentication
  PagerDuty rate limits: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUx-rate-limiting
  PagerDuty Status: https://status.pagerduty.com/

CONFIGURATION
  cutoff_days (integer): Fetch pipelines updated in the last number of days
  default_severity (string): A default severity category if not present
  exclude_services (array): List of PagerDuty service names to ignore incidents from. If not set, all incidents will be pulled.
  incident_log_entries_overview (boolean): If true, will return a subset of log entries that show only the most important changes to the incident.
  max_retries (integer): Maximum number of PagerDuty API request retries to perform upon connection errors. The source will pause for an exponentially increasing number of seconds before retrying.
  page_size (integer): page size to use when querying PagerDuty API
  service_details (array): List of PagerDuty service additional details to include.
  token (string) required secret: API key for PagerDuty API authentication
  secret fields: token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pagerduty

  # Inspect as JSON
  pm connectors inspect source-pagerduty --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  PagerDuty API reference: https://developer.pagerduty.com/api-reference/
  PagerDuty authentication: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUw-authentication
  PagerDuty rate limits: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUx-rate-limiting
  PagerDuty Status: https://status.pagerduty.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
