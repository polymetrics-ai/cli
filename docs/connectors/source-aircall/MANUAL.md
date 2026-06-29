# pm connectors inspect source-aircall

```text
NAME
  pm connectors inspect source-aircall - Aircall connector manual

SYNOPSIS
  pm connectors inspect source-aircall
  pm connectors inspect source-aircall --json
  pm credentials add <name> --connector source-aircall [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Aircall catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/aircall.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.aircall.io/api-references/

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
  API documentation: https://developer.aircall.io/api-references/
  Authentication: https://developer.aircall.io/api-references/#authentication
  Rate limits: https://developer.aircall.io/api-references/#rate-limit

CONFIGURATION
  api_id (string) required secret: App ID found at settings https://dashboard.aircall.io/integrations/api-keys
  api_token (string) required secret: App token found at settings (Ref- https://dashboard.aircall.io/integrations/api-keys)
  start_date (string) required: Date time filter for incremental filter, Specify which date to extract from.
  secret fields: api_id, api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-aircall

  # Inspect as JSON
  pm connectors inspect source-aircall --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  API documentation: https://developer.aircall.io/api-references/
  Authentication: https://developer.aircall.io/api-references/#authentication
  Rate limits: https://developer.aircall.io/api-references/#rate-limit

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
