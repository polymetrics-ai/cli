# pm connectors inspect destination-pubsub

```text
NAME
  pm connectors inspect destination-pubsub - Google PubSub connector manual

SYNOPSIS
  pm connectors inspect destination-pubsub
  pm connectors inspect destination-pubsub --json
  pm credentials add <name> --connector destination-pubsub [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google PubSub catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/googlepubsub.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
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
  family: destination_writer
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  Google PubSub documentation: https://cloud.google.com/pubsub/docs

CONFIGURATION
  batching_delay_threshold (integer): Number of ms before the buffer is flushed
  batching_element_count_threshold (integer): Number of messages before the buffer is flushed
  batching_enabled (boolean) required: If TRUE messages will be buffered instead of sending them one by one
  batching_request_bytes_threshold (integer): Number of bytes before the buffer is flushed
  credentials_json (string) required secret: manual intervention needed
  ordering_enabled (boolean) required: If TRUE PubSub publisher will have <a href="https://cloud.google.com/pubsub/docs/ordering">message ordering</a> enabled. Every message will have an ordering key of stream
  project_id (string) required: The GCP project ID for the project containing the target PubSub.
  topic_id (string) required: The PubSub topic ID in the given GCP project ID.
  secret fields: credentials_json

SYNC MODES
  supported sync modes: append
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-pubsub

  # Inspect as JSON
  pm connectors inspect destination-pubsub --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google PubSub documentation: https://cloud.google.com/pubsub/docs

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
