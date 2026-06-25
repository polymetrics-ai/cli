# pm connectors inspect destination-gcs-data-lake

```text
NAME
  pm connectors inspect destination-gcs-data-lake - GCS Data Lake connector manual

SYNOPSIS
  pm connectors inspect destination-gcs-data-lake
  pm connectors inspect destination-gcs-data-lake --json
  pm credentials add <name> --connector destination-gcs-data-lake [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  GCS Data Lake catalog connector for https://docs.airbyte.com/integrations/destinations/gcs-data-lake. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-gcs-data-lake:1.0.10 (metadata only; not executed)

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
  Cloud Storage documentation: https://cloud.google.com/storage/docs
  Service account authentication: https://cloud.google.com/iam/docs/service-accounts
  Access control: https://cloud.google.com/storage/docs/access-control
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/gcs-data-lake

CONFIGURATION
  catalog_type (object) required: Specifies the type of Iceberg catalog (BigLake or Polaris).
  gcp_location (string) required: The GCP location (region) for BigLake metastore resources. For example: "us-central1" or "us". See <a href="https://cloud.google.com/biglake/docs/locations">BigLake locations</a...
  gcp_project_id (string): The GCP project ID where resources are located. If not specified, it will be extracted from the service account credentials.
  gcs_bucket_name (string) required: The name of the GCS bucket that will host the Iceberg data.
  gcs_endpoint (string): Optional custom GCS endpoint URL. Use this for testing with local GCS emulators.
  main_branch_name (string) required: The primary or default branch name in the catalog. Most query engines will use "main" by default. See <a href="https://iceberg.apache.org/docs/latest/branching/">Iceberg documen...
  namespace (string) required: The default namespace to use for tables. This will ONLY be used if the `Destination Namespace` setting is set to `Destination-defined` or `Source-defined`
  service_account_json (string) required secret: The contents of the JSON service account key file. See the <a href="https://cloud.google.com/iam/docs/creating-managing-service-account-keys">Google Cloud documentation</a> for ...
  warehouse_location (string) required: The root location of the data warehouse used by the Iceberg catalog. Must include the storage protocol "gs://" for Google Cloud Storage. For example: "gs://your-bucket/path/to/w...
  secret fields: catalog_type.client_id, catalog_type.client_secret, service_account_json

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/gcs-data-lake

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-gcs-data-lake

  # Inspect as JSON
  pm connectors inspect destination-gcs-data-lake --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  GCS Data Lake documentation: https://docs.airbyte.com/integrations/destinations/gcs-data-lake

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
