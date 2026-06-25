---
name: pm-destination-s3
description: S3 connector knowledge and safe action guide.
---

# pm-destination-s3

## Purpose

S3 catalog connector for https://docs.airbyte.com/integrations/destinations/s3. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-s3:1.9.8 (metadata only; not executed)

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
- priority_wave: 1
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- AWS S3 documentation: https://docs.aws.amazon.com/s3/
- IAM authentication: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
- Bucket policies and permissions: https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-policy-language-overview.html
- Request rate and performance: https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html
- AWS Service Health Dashboard: https://health.aws.amazon.com/health/status
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/s3

## Configuration

- access_key_id (string) secret: The access key ID to access the S3 bucket. Airbyte requires Read and Write permissions to the given bucket. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-...
- file_name_pattern (string): Pattern to match file names in the bucket directory. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/ListingKeysUsingAPIs.html">here</a>
- format (object) required: Format of the data output. See <a href="https://docs.airbyte.com/integrations/destinations/s3/#supported-output-schema">here</a> for more details
- role_arn (string): The ARN of the AWS role to assume. Only usable in Airbyte Cloud.
- s3_bucket_name (string) required: The name of the S3 bucket. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html">here</a>.
- s3_bucket_path (string) required: Directory under the S3 bucket where data will be written. Read more <a href="https://docs.airbyte.com/integrations/destinations/s3#:~:text=to%20format%20the-,bucket%20path,-%3A"...
- s3_bucket_region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
- s3_endpoint (string): Your S3 endpoint url. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/s3.html#:~:text=Service%20endpoints-,Amazon%20S3%20endpoints,-When%20you%20use">here</a>
- s3_path_format (string): Format string on how data will be organized inside the bucket directory. Read more <a href="https://docs.airbyte.com/integrations/destinations/s3#:~:text=The%20full%20path%20of%...
- secret_access_key (string) secret: The corresponding secret to the access key ID. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys">here</a>
- secret fields: access_key_id, secret_access_key

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/s3

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-s3
```

### Inspect as JSON

```bash
pm connectors inspect destination-s3 --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [S3 documentation](https://docs.airbyte.com/integrations/destinations/s3)
