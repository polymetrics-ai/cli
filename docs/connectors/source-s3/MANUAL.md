# pm connectors inspect source-s3

```text
NAME
  pm connectors inspect source-s3 - S3 connector manual

SYNOPSIS
  pm connectors inspect source-s3
  pm connectors inspect source-s3 --json
  pm credentials add <name> --connector source-s3 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  S3 catalog connector for https://docs.airbyte.com/integrations/sources/s3. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: file_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-s3:4.15.10 (metadata only; not executed)

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
  family: file_object_source
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Changelog: https://docs.aws.amazon.com/AmazonS3/latest/userguide/WhatsNew.html
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/s3

CONFIGURATION
  aws_access_key_id (string) secret: In order to access private Buckets stored on AWS S3, this connector requires credentials with the proper permissions. If accessing publicly available data, this field is not nec...
  aws_secret_access_key (string) secret: In order to access private Buckets stored on AWS S3, this connector requires credentials with the proper permissions. If accessing publicly available data, this field is not nec...
  bucket (string) required: Name of the S3 bucket where the file(s) exist.
  dataset (string): Deprecated and will be removed soon. Please do not use this field anymore and use streams.name instead. The name of the stream you would like this source to output. Can contain ...
  delivery_method (object)
  endpoint (string): Endpoint to an S3 compatible service. Leave empty to use AWS.
  format (object): Deprecated and will be removed soon. Please do not use this field anymore and use streams.format instead. The format of the files you'd like to replicate
  path_pattern (string): Deprecated and will be removed soon. Please do not use this field anymore and use streams.globs instead. A regular expression which tells the connector which files to replicate....
  provider (object): Deprecated and will be removed soon. Please do not use this field anymore and use bucket, aws_access_key_id, aws_secret_access_key and endpoint instead. Use this to load files f...
  region_name (string): AWS region where the S3 bucket is located. If not provided, the region will be determined automatically.
  role_arn (string): Specifies the Amazon Resource Name (ARN) of an IAM role that you want to use to perform operations requested using this profile. Set the External ID to the Airbyte workspace ID,...
  schema (string): Deprecated and will be removed soon. Please do not use this field anymore and use streams.input_schema instead. Optionally provide a schema to enforce, as a valid JSON string. E...
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
  streams (array) required: Each instance of this configuration defines a <a href="https://docs.airbyte.com/cloud/core-concepts#stream">stream</a>. Use this to define which files belong in the stream, thei...
  secret fields: aws_access_key_id, aws_secret_access_key, provider.aws_access_key_id, provider.aws_secret_access_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/s3

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-s3

  # Inspect as JSON
  pm connectors inspect source-s3 --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  S3 documentation: https://docs.airbyte.com/integrations/sources/s3

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
