---
name: pm-destination-aws-datalake
description: AWS Datalake connector knowledge and safe action guide.
---

# pm-destination-aws-datalake

## Purpose

AWS Datalake catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/awsdatalake.svg
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

- aws_account_id (string): target aws account id
- bucket_name (string) required: The name of the S3 bucket. Read more <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html">here</a>.
- bucket_prefix (string): S3 prefix
- credentials (object) required: Choose How to Authenticate to AWS.
- format (object): Format of the data output.
- glue_catalog_float_as_decimal (boolean): Cast float/double as decimal(38,18). This can help achieve higher accuracy and represent numbers correctly as received from the source.
- lakeformation_database_default_tag_key (string): Add a default tag key to databases created by this destination
- lakeformation_database_default_tag_values (string): Add default values for the `Tag Key` to databases created by this destination. Comma separate for multiple values.
- lakeformation_database_name (string) required: The default database this destination will use to create tables in per stream. Can be changed per connection by customizing the namespace.
- lakeformation_governed_tables (boolean): Whether to create tables as LF governed tables.
- partitioning (string): Partition data by cursor fields when a cursor field is a date
- region (string) required: The region of the S3 bucket. See <a href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions">here</a> for all ...
- secret fields: credentials.aws_access_key_id, credentials.aws_secret_access_key

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
pm connectors inspect destination-aws-datalake
```

### Inspect as JSON

```bash
pm connectors inspect destination-aws-datalake --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.
