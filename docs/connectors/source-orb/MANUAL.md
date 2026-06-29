# pm connectors inspect source-orb

```text
NAME
  pm connectors inspect source-orb - Orb connector manual

SYNOPSIS
  pm connectors inspect source-orb
  pm connectors inspect source-orb --json
  pm credentials add <name> --connector source-orb [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Orb catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/orb.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.withorb.com/reference/api-reference

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
  Orb API reference: https://docs.withorb.com/reference/api-reference
  Orb authentication: https://docs.withorb.com/reference/authentication
  Orb Status: https://status.withorb.com/

CONFIGURATION
  api_key (string) required secret: Orb API Key, issued from the Orb admin console.
  end_date (string): UTC date and time in the format 2022-03-01T00:00:00Z. Any data with created_at after this data will not be synced. For Subscription Usage, this becomes the `timeframe_start` API...
  lookback_window_days (integer): When set to N, the connector will always refresh resources created within the past N days. By default, updated objects that are not newly created are not incrementally synced.
  numeric_event_properties_keys (array): Property key names to extract from all events, in order to enrich ledger entries corresponding to an event deduction.
  plan_id (string): Orb Plan ID to filter subscriptions that should have usage fetched.
  start_date (string) required: UTC date and time in the format 2022-03-01T00:00:00Z. Any data with created_at before this data will not be synced. For Subscription Usage, this becomes the `timeframe_start` AP...
  string_event_properties_keys (array): Property key names to extract from all events, in order to enrich ledger entries corresponding to an event deduction.
  subscription_usage_grouping_key (string): Property key name to group subscription usage by.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-orb

  # Inspect as JSON
  pm connectors inspect source-orb --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Orb API reference: https://docs.withorb.com/reference/api-reference
  Orb authentication: https://docs.withorb.com/reference/authentication
  Orb Status: https://status.withorb.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
