# pm connectors inspect destination-milvus

```text
NAME
  pm connectors inspect destination-milvus - Milvus connector manual

SYNOPSIS
  pm connectors inspect destination-milvus
  pm connectors inspect destination-milvus --json
  pm credentials add <name> --connector destination-milvus [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Milvus catalog connector for https://docs.airbyte.com/integrations/destinations/milvus. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-milvus:0.0.58 (metadata only; not executed)

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
  Milvus documentation: https://milvus.io/docs
  Authentication: https://milvus.io/docs/authenticate.md
  Release notes: https://milvus.io/docs/release_notes.md
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/milvus

CONFIGURATION
  embedding (object) required: Embedding configuration
  indexing (object) required: Indexing configuration
  omit_raw_text (boolean): Do not store the text that gets embedded along with the vector and the metadata in the destination. If set to true, only the vector and the metadata will be stored - in this cas...
  processing (object) required
  secret fields: embedding.api_key, embedding.cohere_key, embedding.openai_key, indexing.auth.password, indexing.auth.token

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/milvus

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-milvus

  # Inspect as JSON
  pm connectors inspect destination-milvus --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Milvus documentation: https://docs.airbyte.com/integrations/destinations/milvus

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
