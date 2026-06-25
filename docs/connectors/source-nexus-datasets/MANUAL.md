# pm connectors inspect source-nexus-datasets

```text
NAME
  pm connectors inspect source-nexus-datasets - Infor Nexus connector manual

SYNOPSIS
  pm connectors inspect source-nexus-datasets
  pm connectors inspect source-nexus-datasets --json
  pm credentials add <name> --connector source-nexus-datasets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Infor Nexus catalog connector for https://docs.airbyte.com/integrations/sources/nexus-datasets. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-nexus-datasets:0.1.9 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/nexus-datasets

CONFIGURATION
  access_key_id (string) required: The access key ID for the Infor Nexus API.
  api_key (string) required: The Infor Data API Key.
  base_url (string) required: The base URL of the Infor Nexus datasets API.
  dataset_name (string) required: The name of the dataset to export.
  mode (string): Infor Streaming mode (Full or Incremental). https://developer.infornexus.com/api/version-3.1-api-methods
  secret_key (string) required: The secret key for HMAC authentication.
  user_id (string) required: The user ID for the Infor Nexus API.

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/nexus-datasets

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-nexus-datasets

  # Inspect as JSON
  pm connectors inspect source-nexus-datasets --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Infor Nexus documentation: https://docs.airbyte.com/integrations/sources/nexus-datasets

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
