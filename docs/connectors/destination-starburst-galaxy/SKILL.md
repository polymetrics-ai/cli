---
name: pm-destination-starburst-galaxy
description: Starburst Galaxy connector knowledge and safe action guide.
---

# pm-destination-starburst-galaxy

## Purpose

Starburst Galaxy catalog connector for https://docs.airbyte.com/integrations/destinations/starburst-galaxy. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-starburst-galaxy:0.0.1 (metadata only; not executed)

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

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/starburst-galaxy

## Configuration

- accept_terms (boolean) required: You must agree to the Starburst Galaxy <a href="https://www.starburst.io/terms/">terms & conditions</a> to use this connector.
- catalog (string) required: Name of the Starburst Galaxy Amazon S3 catalog.
- catalog_schema (string): The default Starburst Galaxy Amazon S3 catalog schema where tables are written to if the source does not specify a namespace. Defaults to "public".
- password (string) required secret: Starburst Galaxy password for the specified user.
- port (string): Starburst Galaxy cluster port.
- purge_staging_table (boolean): Defaults to 'true'. Switch to 'false' for debugging purposes.
- server_hostname (string) required: Starburst Galaxy cluster hostname.
- staging_object_store (object) required: Temporary storage on which temporary Iceberg table is created.
- username (string) required: Starburst Galaxy user.
- secret fields: password, staging_object_store.s3_access_key_id, staging_object_store.s3_secret_access_key

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/starburst-galaxy

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-starburst-galaxy
```

### Inspect as JSON

```bash
pm connectors inspect destination-starburst-galaxy --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Starburst Galaxy documentation](https://docs.airbyte.com/integrations/destinations/starburst-galaxy)
