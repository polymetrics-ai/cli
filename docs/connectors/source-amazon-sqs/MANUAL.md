# pm connectors inspect source-amazon-sqs

```text
NAME
  pm connectors inspect source-amazon-sqs - Amazon SQS connector manual

SYNOPSIS
  pm connectors inspect source-amazon-sqs
  pm connectors inspect source-amazon-sqs --json
  pm credentials add <name> --connector source-amazon-sqs [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Amazon SQS catalog connector for https://docs.airbyte.com/integrations/sources/amazon-sqs. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-amazon-sqs:1.0.15 (metadata only; not executed)

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
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/amazon-sqs

CONFIGURATION
  access_key (string) required secret: The Access Key ID of the AWS IAM Role to use for pulling messages
  attributes_to_return (string): Comma separated list of Mesage Attribute names to return
  max_batch_size (integer): Max amount of messages to get in one batch (10 max)
  max_wait_time (integer): Max amount of time in seconds to wait for messages in a single poll (20 max)
  queue_url (string) required: URL of the SQS Queue
  region (string) required: AWS Region of the SQS Queue
  secret_key (string) required secret: The Secret Key of the AWS IAM Role to use for pulling messages
  target (string): Note - Different targets have different attribute enum requirements, please refer actions sections in https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/Welco...
  visibility_timeout (integer): Modify the Visibility Timeout of the individual message from the Queue's default (seconds).
  secret fields: access_key, secret_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/amazon-sqs

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-amazon-sqs

  # Inspect as JSON
  pm connectors inspect source-amazon-sqs --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Amazon SQS documentation: https://docs.airbyte.com/integrations/sources/amazon-sqs

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
