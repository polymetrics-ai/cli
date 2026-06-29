---
name: pm-destination-dynamodb
description: DynamoDB connector knowledge and safe action guide.
---

# pm-destination-dynamodb

## Purpose

DynamoDB catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/dynamodb.svg
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

- DynamoDB documentation: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html

## Configuration

- access_key_id (string) required secret: manual intervention needed
- dynamodb_endpoint (string): This is your DynamoDB endpoint url.(if you are working with AWS DynamoDB, just leave empty).
- dynamodb_region (string) required: The region of the DynamoDB.
- dynamodb_table_name_prefix (string) required: The prefix to use when naming DynamoDB tables.
- secret_access_key (string) required secret: The corresponding secret to the access key id.
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
pm connectors inspect destination-dynamodb
```

### Inspect as JSON

```bash
pm connectors inspect destination-dynamodb --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [DynamoDB documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html)
