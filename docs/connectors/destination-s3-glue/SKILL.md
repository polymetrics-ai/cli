---
name: pm-destination-s3-glue
description: S3 Glue connector knowledge and safe action guide.
---

# pm-destination-s3-glue

## Purpose

S3 Glue catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/s3-glue.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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

- manual intervention needed

## Configuration

- access_key_id (string) secret: manual intervention needed
- file_name_pattern (string): The pattern allows you to set the file-name format for the S3 staging file(s)
- format (object) required: manual intervention needed
- glue_database (string) required: Name of the glue database for creating the tables, leave blank if no integration
- glue_serialization_library (string) required: The library that your query engine will use for reading and writing data in your lake.
- s3_bucket_name (string) required: The name of the S3 bucket. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html">here</a>.
- s3_bucket_path (string) required: manual intervention needed
- s3_bucket_region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
- s3_endpoint (string): Your S3 endpoint url. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/s3.html#:~:text=Service%20endpoints-,Amazon%20S3%20endpoints,-When%20you%20use">here</a>
- s3_path_format (string): manual intervention needed
- secret_access_key (string) secret: The corresponding secret to the access key ID. Read more <a href="https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys">here</a>
- secret fields: access_key_id, secret_access_key

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-s3-glue
```

### Inspect as JSON

```bash
pm connectors inspect destination-s3-glue --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.
