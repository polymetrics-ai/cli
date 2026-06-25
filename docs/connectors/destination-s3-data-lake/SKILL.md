---
name: pm-destination-s3-data-lake
description: S3 Data Lake connector knowledge and safe action guide.
---

# pm-destination-s3-data-lake

## Purpose

S3 Data Lake catalog connector for https://docs.airbyte.com/integrations/destinations/s3-data-lake. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-s3-data-lake:0.3.52 (metadata only; not executed)

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: destination_writer
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- AWS S3 documentation: https://docs.aws.amazon.com/s3/
- IAM authentication: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
- Bucket policies and permissions: https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-policy-language-overview.html
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/s3-data-lake

## Configuration

- access_key_id (string) secret: The AWS Access Key ID with permissions for S3 and Glue operations.
- catalog_type (object) required: Specifies the type of Iceberg catalog (e.g., NESSIE, GLUE, REST, POLARIS) and its associated configuration.
- flush_batch_size_mb (integer): The approximate size in megabytes of each batch of data written to Iceberg. Smaller values flush more frequently, improving data freshness and reducing data loss on failure, but...
- main_branch_name (string) required: The primary or default branch name in the catalog. Most query engines will use "main" by default. See <a href="https://iceberg.apache.org/docs/latest/branching/">Iceberg documen...
- s3_bucket_name (string) required: The name of the S3 bucket that will host the Iceberg data.
- s3_bucket_region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
- s3_endpoint (string): Your S3 endpoint url. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/s3.html#:~:text=Service%20endpoints-,Amazon%20S3%20endpoints,-When%20you%20use">here</a>
- secret_access_key (string) secret: The AWS Secret Access Key paired with the Access Key ID for AWS authentication.
- warehouse_location (string) required: The root location of the data warehouse used by the Iceberg catalog. Typically includes a bucket name and path within that bucket. For AWS Glue and Nessie, must include the stor...
- secret fields: access_key_id, catalog_type.access_token, catalog_type.client_id, catalog_type.client_secret, secret_access_key

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/s3-data-lake

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-s3-data-lake
```

### Inspect as JSON

```bash
pm connectors inspect destination-s3-data-lake --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [S3 Data Lake documentation](https://docs.airbyte.com/integrations/destinations/s3-data-lake)
