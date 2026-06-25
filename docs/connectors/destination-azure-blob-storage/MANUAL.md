# pm connectors inspect destination-azure-blob-storage

```text
NAME
  pm connectors inspect destination-azure-blob-storage - Azure Blob Storage connector manual

SYNOPSIS
  pm connectors inspect destination-azure-blob-storage
  pm connectors inspect destination-azure-blob-storage --json
  pm credentials add <name> --connector destination-azure-blob-storage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Azure Blob Storage catalog connector for https://docs.airbyte.com/integrations/destinations/azure-blob-storage. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-azure-blob-storage:1.1.7 (metadata only; not executed)

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
  Azure Blob Storage documentation: https://learn.microsoft.com/en-us/azure/storage/blobs/
  Authorize access: https://learn.microsoft.com/en-us/azure/storage/common/authorize-data-access
  Access control: https://learn.microsoft.com/en-us/azure/storage/blobs/data-lake-storage-access-control
  Scalability and performance: https://learn.microsoft.com/en-us/azure/storage/blobs/scalability-targets
  Azure Status: https://status.azure.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/azure-blob-storage

CONFIGURATION
  azure_blob_storage_account_key (string) secret: The Azure Blob Storage account key. If you set this value, you must not set the "Shared Access Signature", "Azure Tenant ID", "Azure Client ID", or "Azure Client Secret" fields.
  azure_blob_storage_account_name (string) required: The name of the Azure Blob Storage Account. Read more <a href="https://learn.microsoft.com/en-gb/azure/storage/blobs/storage-blobs-introduction#storage-accounts">here</a>.
  azure_blob_storage_container_name (string) required: The name of the Azure Blob Storage Container. Read more <a href="https://learn.microsoft.com/en-gb/azure/storage/blobs/storage-blobs-introduction#containers">here</a>.
  azure_blob_storage_endpoint_domain_name (string): This is Azure Blob Storage endpoint domain name. Leave default value (or leave it empty if run container from command line) to use Microsoft native from example.
  azure_blob_storage_spill_size (integer): The amount of megabytes after which the connector should spill the records in a new blob object. Make sure to configure size greater than individual records. Enter 0 if not appl...
  azure_client_id (string): The Azure Active Directory (Entra ID) client ID. Required for Entra ID authentication.
  azure_client_secret (string) secret: The Azure Active Directory (Entra ID) client secret. Required for Entra ID authentication.
  azure_tenant_id (string): The Azure Active Directory (Entra ID) tenant ID. Required for Entra ID authentication.
  format (object) required: Format of the data output.
  shared_access_signature (string) secret: A shared access signature (SAS) provides secure delegated access to resources in your storage account. Read more <a href="https://learn.microsoft.com/en-gb/azure/storage/common/...
  secret fields: azure_blob_storage_account_key, azure_client_secret, shared_access_signature

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/azure-blob-storage

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-azure-blob-storage

  # Inspect as JSON
  pm connectors inspect destination-azure-blob-storage --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Azure Blob Storage documentation: https://docs.airbyte.com/integrations/destinations/azure-blob-storage

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
