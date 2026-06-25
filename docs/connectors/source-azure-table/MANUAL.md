# pm connectors inspect source-azure-table

```text
NAME
  pm connectors inspect source-azure-table - Azure Table Storage connector manual

SYNOPSIS
  pm connectors inspect source-azure-table
  pm connectors inspect source-azure-table --json
  pm credentials add <name> --connector source-azure-table [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Azure Table Storage catalog connector for https://docs.airbyte.com/integrations/sources/azure-table. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-azure-table:0.1.57 (metadata only; not executed)

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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/azure-table

CONFIGURATION
  storage_access_key (string) required secret: Azure Table Storage Access Key. See the <a href="https://docs.airbyte.com/integrations/sources/azure-table">docs</a> for more information on how to obtain this key.
  storage_account_name (string) required: The name of your storage account.
  storage_endpoint_suffix (string): Azure Table Storage service account URL suffix. See the <a href="https://docs.airbyte.com/integrations/sources/azure-table">docs</a> for more information on how to obtain endpoi...
  secret fields: storage_access_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/azure-table

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-azure-table

  # Inspect as JSON
  pm connectors inspect source-azure-table --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Azure Table Storage documentation: https://docs.airbyte.com/integrations/sources/azure-table

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
