---
name: pm-source-microsoft-sharepoint
description: Microsoft SharePoint connector knowledge and safe action guide.
---

# pm-source-microsoft-sharepoint

## Purpose

Microsoft SharePoint catalog connector for https://docs.airbyte.com/integrations/sources/microsoft-sharepoint. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: file_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-microsoft-sharepoint:0.10.4 (metadata only; not executed)

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
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- SharePoint REST API: https://learn.microsoft.com/en-us/sharepoint/dev/sp-add-ins/get-to-know-the-sharepoint-rest-service
- SharePoint authentication: https://learn.microsoft.com/en-us/sharepoint/dev/sp-add-ins/authorization-and-authentication-of-sharepoint-add-ins
- Microsoft 365 Status: https://status.office365.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/microsoft-sharepoint

## Configuration

- credentials (object) required: Credentials for connecting to the One Drive API
- delivery_method (object)
- folder_path (string): Path to a specific folder within the drives to search for files. Leave empty to search all folders of the drives. This does not apply to shared items.
- search_scope (string): Specifies the location(s) to search for files. Valid options are 'ACCESSIBLE_DRIVES' for all SharePoint drives the user can access, 'SHARED_ITEMS' for shared items the user has ...
- site_url (string): Url of SharePoint site to search for files. Leave empty to search in the main site. Use 'https://<tenant_name>.sharepoint.com/sites/' to iterate over all sites.
- start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
- streams (array) required: Each instance of this configuration defines a <a href="https://docs.airbyte.com/cloud/core-concepts#stream">stream</a>. Use this to define which files belong in the stream, thei...
- secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.tenant_id, credentials.user_principal_name

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/microsoft-sharepoint

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-microsoft-sharepoint
```

### Inspect as JSON

```bash
pm connectors inspect source-microsoft-sharepoint --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Microsoft SharePoint documentation](https://docs.airbyte.com/integrations/sources/microsoft-sharepoint)
