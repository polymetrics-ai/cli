# pm connectors inspect source-paystack

```text
NAME
  pm connectors inspect source-paystack - Paystack connector manual

SYNOPSIS
  pm connectors inspect source-paystack
  pm connectors inspect source-paystack --json
  pm credentials add <name> --connector source-paystack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Paystack catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/paystack.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://paystack.com/docs/api/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Paystack API reference: https://paystack.com/docs/api/
  Paystack authentication: https://paystack.com/docs/api/#authentication
  Paystack rate limits: https://paystack.com/docs/api/#rate-limiting

CONFIGURATION
  lookback_window_days (integer): When set, the connector will always reload data from the past N days, where N is the value set here. This is useful if your data is updated after creation.
  secret_key (string) required secret: The Paystack API key (usually starts with 'sk_live_'; find yours <a href="https://dashboard.paystack.com/#/settings/developer">here</a>).
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: secret_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-paystack

  # Inspect as JSON
  pm connectors inspect source-paystack --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Paystack API reference: https://paystack.com/docs/api/
  Paystack authentication: https://paystack.com/docs/api/#authentication
  Paystack rate limits: https://paystack.com/docs/api/#rate-limiting

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
