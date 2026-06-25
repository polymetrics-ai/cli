# pm connectors inspect destination-s3-glue

```text
NAME
  pm connectors inspect destination-s3-glue - S3 Glue connector manual

SYNOPSIS
  pm connectors inspect destination-s3-glue
  pm connectors inspect destination-s3-glue --json
  pm credentials add <name> --connector destination-s3-glue [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  S3 Glue catalog connector for https://docs.airbyte.com/integrations/destinations/s3-glue. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-s3-glue:0.1.11 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/s3-glue

CONFIGURATION
  access_key_id (string) secret: The access key ID to access the S3 bucket. Airbyte requires Read and Write permissions to the given bucket. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-...
  file_name_pattern (string): The pattern allows you to set the file-name format for the S3 staging file(s)
  format (object) required: Format of the data output. See <a href="https://docs.airbyte.com/integrations/destinations/s3/#supported-output-schema">here</a> for more details
  glue_database (string) required: Name of the glue database for creating the tables, leave blank if no integration
  glue_serialization_library (string) required: The library that your query engine will use for reading and writing data in your lake.
  s3_bucket_name (string) required: The name of the S3 bucket. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html">here</a>.
  s3_bucket_path (string) required: Directory under the S3 bucket where data will be written. Read more <a href="https://docs.airbyte.com/integrations/destinations/s3#:~:text=to%20format%20the-,bucket%20path,-%3A"...
  s3_bucket_region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
  s3_endpoint (string): Your S3 endpoint url. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/s3.html#:~:text=Service%20endpoints-,Amazon%20S3%20endpoints,-When%20you%20use">here</a>
  s3_path_format (string): Format string on how data will be organized inside the S3 bucket directory. Read more <a href="https://docs.airbyte.com/integrations/destinations/s3#:~:text=The%20full%20path%20...
  secret_access_key (string) secret: The corresponding secret to the access key ID. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys">here</a>
  secret fields: access_key_id, secret_access_key

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/s3-glue

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-s3-glue

  # Inspect as JSON
  pm connectors inspect destination-s3-glue --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  S3 Glue documentation: https://docs.airbyte.com/integrations/destinations/s3-glue

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
