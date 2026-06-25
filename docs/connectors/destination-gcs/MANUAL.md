# pm connectors inspect destination-gcs

```text
NAME
  pm connectors inspect destination-gcs - Google Cloud Storage (GCS) connector manual

SYNOPSIS
  pm connectors inspect destination-gcs
  pm connectors inspect destination-gcs --json
  pm credentials add <name> --connector destination-gcs [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Cloud Storage (GCS) catalog connector for https://docs.airbyte.com/integrations/destinations/gcs. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-gcs:0.4.9 (metadata only; not executed)

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
  priority_wave: 2
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  Cloud Storage documentation: https://cloud.google.com/storage/docs
  Service account authentication: https://cloud.google.com/iam/docs/service-accounts
  Access control: https://cloud.google.com/storage/docs/access-control
  Google Cloud Release Notes: https://cloud.google.com/release-notes
  Quotas and limits: https://cloud.google.com/storage/quotas
  Google Cloud Status: https://status.cloud.google.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/gcs

CONFIGURATION
  credential (object) required: An HMAC key is a type of credential and can be associated with a service account or a user account in Cloud Storage. Read more <a href="https://cloud.google.com/storage/docs/aut...
  format (object) required: Output data format. One of the following formats must be selected - <a href="https://cloud.google.com/bigquery/docs/loading-data-cloud-storage-avro#advantages_of_avro">AVRO</a> ...
  gcs_bucket_name (string) required: You can find the bucket name in the App Engine Admin console Application Settings page, under the label Google Cloud Storage Bucket. Read more <a href="https://cloud.google.com/...
  gcs_bucket_path (string) required: GCS Bucket Path string Subdirectory under the above bucket to sync the data into.
  gcs_bucket_region (string): Select a Region of the GCS Bucket. Read more <a href="https://cloud.google.com/storage/docs/locations">here</a>.
  secret fields: credential.hmac_key_access_id, credential.hmac_key_secret

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/gcs

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-gcs

  # Inspect as JSON
  pm connectors inspect destination-gcs --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Cloud Storage (GCS) documentation: https://docs.airbyte.com/integrations/destinations/gcs

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
