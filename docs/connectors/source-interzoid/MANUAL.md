# pm connectors inspect source-interzoid

```text
NAME
  pm connectors inspect source-interzoid - Interzoid connector manual

SYNOPSIS
  pm connectors inspect source-interzoid
  pm connectors inspect source-interzoid --json
  pm credentials add <name> --connector source-interzoid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Interzoid catalog connector for https://docs.airbyte.com/integrations/sources/interzoid. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-interzoid:0.0.6 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/interzoid

CONFIGURATION
  address (string) required: Street address for the Street Address Matching stream.
  address_match_algorithm (string) required: Algorithm for address matching (e.g. model-v3-narrow).
  api_key (string) required secret: Your Interzoid API key.
  company (string) required: Company name for the Company Name Matching stream.
  company_match_algorithm (string) required: Algorithm for company name matching (e.g. model-v4-wide).
  fullname (string) required: Full name for the Individual Name Matching stream.
  org (string) required: Organization name for the Standardize Company Names stream.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/interzoid

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-interzoid

  # Inspect as JSON
  pm connectors inspect source-interzoid --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Interzoid documentation: https://docs.airbyte.com/integrations/sources/interzoid

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
