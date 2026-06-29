# pm connectors inspect destination-s3

```text
NAME
  pm connectors inspect destination-s3 - S3 connector manual

SYNOPSIS
  pm connectors inspect destination-s3
  pm connectors inspect destination-s3 --json
  pm credentials add <name> --connector destination-s3 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  S3 catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/s3.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.aws.amazon.com/s3/

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
  support level: certified

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
  priority_wave: 1
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  AWS S3 documentation: https://docs.aws.amazon.com/s3/
  IAM authentication: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
  Bucket policies and permissions: https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-policy-language-overview.html
  Request rate and performance: https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html
  AWS Service Health Dashboard: https://health.aws.amazon.com/health/status

CONFIGURATION
  access_key_id (string) secret: manual intervention needed
  file_name_pattern (string): Pattern to match file names in the bucket directory. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/ListingKeysUsingAPIs.html">here</a>
  format (object) required: manual intervention needed
  role_arn (string): manual intervention needed
  s3_bucket_name (string) required: The name of the S3 bucket. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html">here</a>.
  s3_bucket_path (string) required: manual intervention needed
  s3_bucket_region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
  s3_endpoint (string): Your S3 endpoint url. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/s3.html#:~:text=Service%20endpoints-,Amazon%20S3%20endpoints,-When%20you%20use">here</a>
  s3_path_format (string): manual intervention needed
  secret_access_key (string) secret: The corresponding secret to the access key ID. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys">here</a>
  secret fields: access_key_id, secret_access_key

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-s3

  # Inspect as JSON
  pm connectors inspect destination-s3 --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  AWS S3 documentation: https://docs.aws.amazon.com/s3/
  IAM authentication: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
  Bucket policies and permissions: https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-policy-language-overview.html
  Request rate and performance: https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html
  AWS Service Health Dashboard: https://health.aws.amazon.com/health/status

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
