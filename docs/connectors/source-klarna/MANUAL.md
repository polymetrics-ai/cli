# pm connectors inspect source-klarna

```text
NAME
  pm connectors inspect source-klarna - Klarna connector manual

SYNOPSIS
  pm connectors inspect source-klarna
  pm connectors inspect source-klarna --json
  pm credentials add <name> --connector source-klarna [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Klarna catalog connector for https://docs.airbyte.com/integrations/sources/klarna. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-klarna:0.3.23 (metadata only; not executed)

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
  Klarna API reference: https://docs.klarna.com/api/
  Klarna authentication: https://docs.klarna.com/api/authentication/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/klarna

CONFIGURATION
  password (string) required secret: A string which is associated with your Merchant ID and is used to authorize use of Klarna's APIs (https://developers.klarna.com/api/#authentication)
  playground (boolean) required: Propertie defining if connector is used against playground or production environment
  region (string) required: Base url region (For playground eu https://docs.klarna.com/klarna-payments/api/payments-api/#tag/API-URLs). Supported 'eu', 'na', 'oc'
  username (string) required: Consists of your Merchant ID (eid) - a unique number that identifies your e-store, combined with a random string (https://developers.klarna.com/api/#authentication)
  secret fields: password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/klarna

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-klarna

  # Inspect as JSON
  pm connectors inspect source-klarna --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Klarna documentation: https://docs.airbyte.com/integrations/sources/klarna

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
