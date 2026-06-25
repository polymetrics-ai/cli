---
name: pm-source-aws-cloudtrail
description: Aws Cloudtrail connector knowledge and safe action guide.
---

# pm-source-aws-cloudtrail

## Purpose

Aws Cloudtrail catalog connector for https://docs.airbyte.com/integrations/sources/aws-cloudtrail. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-aws-cloudtrail:1.1.0 (metadata only; not executed)

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

- family: declarative_http_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/aws-cloudtrail

## Configuration

- aws_key_id (string) required secret: AWS CloudTrail Access Key ID. See the <a href="https://docs.airbyte.com/integrations/sources/aws-cloudtrail">docs</a> for more information on how to obtain this key.
- aws_region_name (string) required: The default AWS Region to use, for example, us-west-1 or us-west-2. When specifying a Region inline during client initialization, this property is named region_name.
- aws_secret_key (string) required secret: AWS CloudTrail Access Key ID. See the <a href="https://docs.airbyte.com/integrations/sources/aws-cloudtrail">docs</a> for more information on how to obtain this key.
- lookup_attributes_filter (object)
- start_date (string): The date you would like to replicate data. Data in AWS CloudTrail is available for last 90 days only. Format: YYYY-MM-DD.
- secret fields: aws_key_id, aws_secret_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/aws-cloudtrail

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-aws-cloudtrail
```

### Inspect as JSON

```bash
pm connectors inspect source-aws-cloudtrail --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Aws Cloudtrail documentation](https://docs.airbyte.com/integrations/sources/aws-cloudtrail)
