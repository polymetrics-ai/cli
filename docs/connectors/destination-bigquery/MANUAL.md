# pm connectors inspect destination-bigquery

```text
NAME
  pm connectors inspect destination-bigquery - BigQuery connector manual

SYNOPSIS
  pm connectors inspect destination-bigquery
  pm connectors inspect destination-bigquery --json
  pm credentials add <name> --connector destination-bigquery [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  BigQuery catalog connector for https://docs.airbyte.com/integrations/destinations/bigquery. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-bigquery:3.0.19 (metadata only; not executed)

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
  Standard SQL reference: https://cloud.google.com/bigquery/docs/reference/standard-sql
  Service account authentication: https://cloud.google.com/iam/docs/service-accounts
  Access control and permissions: https://cloud.google.com/bigquery/docs/access-control
  Release notes: https://cloud.google.com/bigquery/docs/release-notes
  Quotas and limits: https://cloud.google.com/bigquery/quotas
  Google Cloud Status: https://status.cloud.google.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/bigquery

CONFIGURATION
  cdc_deletion_mode (string): Whether to execute CDC deletions as hard deletes (i.e. propagate source deletions to the destination), or soft deletes (i.e. leave a tombstone record in the destination). Defaul...
  credentials_json (string) secret: The contents of the JSON service account key. Check out the <a href="https://docs.airbyte.com/integrations/destinations/bigquery#service-account-key">docs</a> if you need help g...
  dataset_id (string) required: The default BigQuery Dataset ID that tables are replicated to if the source does not specify a namespace. Read more <a href="https://cloud.google.com/bigquery/docs/datasets#crea...
  dataset_location (string) required: The location of the dataset. Warning: Changes made after creation will not be applied. Read more <a href="https://cloud.google.com/bigquery/docs/locations">here</a>.
  disable_type_dedupe (boolean): Write the legacy "raw tables" format, to enable backwards compatibility with older versions of this connector.
  loading_method (object): The way data will be uploaded to BigQuery.
  project_id (string) required: The GCP project ID for the project containing the target BigQuery dataset. Read more <a href="https://cloud.google.com/resource-manager/docs/creating-managing-projects#identifyi...
  raw_data_dataset (string): Airbyte will use this dataset for various internal tables. In legacy raw tables mode, the raw tables will be stored in this dataset. Defaults to "airbyte_internal".
  secret fields: credentials_json, loading_method.credential.hmac_key_access_id, loading_method.credential.hmac_key_secret

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/bigquery

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-bigquery

  # Inspect as JSON
  pm connectors inspect destination-bigquery --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  BigQuery documentation: https://docs.airbyte.com/integrations/destinations/bigquery

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
