---
name: pm-source-azure-blob-storage
description: Azure Blob Storage connector knowledge and safe action guide.
---

# pm-source-azure-blob-storage

## Purpose

Azure Blob Storage catalog connector for https://docs.airbyte.com/integrations/sources/azure-blob-storage. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: file_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-azure-blob-storage:0.8.21 (metadata only; not executed)

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

- family: file_object_source
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- API versions: https://learn.microsoft.com/en-us/rest/api/storageservices/previous-azure-storage-service-versions
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/azure-blob-storage

## Configuration

- azure_blob_storage_account_name (string) required: The account's name of the Azure Blob Storage.
- azure_blob_storage_container_name (string) required: The name of the Azure blob storage container.
- azure_blob_storage_endpoint (string): This is Azure Blob Storage endpoint domain name. Leave default value (or leave it empty if run container from command line) to use Microsoft native from example.
- credentials (object) required: Credentials for connecting to the Azure Blob Storage
- delivery_method (object)
- start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
- streams (array) required: Each instance of this configuration defines a <a href="https://docs.airbyte.com/cloud/core-concepts#stream">stream</a>. Use this to define which files belong in the stream, thei...
- secret fields: credentials.app_client_id, credentials.app_client_secret, credentials.app_tenant_id, credentials.azure_blob_storage_account_key, credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.tenant_id

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/azure-blob-storage

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-azure-blob-storage
```

### Inspect as JSON

```bash
pm connectors inspect source-azure-blob-storage --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Azure Blob Storage documentation](https://docs.airbyte.com/integrations/sources/azure-blob-storage)
