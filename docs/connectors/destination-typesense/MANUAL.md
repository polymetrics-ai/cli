# pm connectors inspect destination-typesense

```text
NAME
  pm connectors inspect destination-typesense - Typesense connector manual

SYNOPSIS
  pm connectors inspect destination-typesense
  pm connectors inspect destination-typesense --json
  pm credentials add <name> --connector destination-typesense [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Typesense catalog connector for https://docs.airbyte.com/integrations/destinations/typesense. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-typesense:0.1.52 (metadata only; not executed)

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
  family: destination_writer
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/typesense

CONFIGURATION
  api_key (string) required: Typesense API Key
  batch_size (integer): How many documents should be imported together. Default 1000
  host (string) required: Hostname of the Typesense instance without protocol. Accept multiple hosts separated by comma.
  path (string): Path of the Typesense instance. Default is none
  port (string): Port of the Typesense instance. Ex: 8108, 80, 443. Default is 8108
  protocol (string): Protocol of the Typesense instance. Ex: http or https. Default is https

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/typesense

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-typesense

  # Inspect as JSON
  pm connectors inspect destination-typesense --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Typesense documentation: https://docs.airbyte.com/integrations/destinations/typesense

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
