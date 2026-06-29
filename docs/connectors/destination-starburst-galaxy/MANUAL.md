# pm connectors inspect destination-starburst-galaxy

```text
NAME
  pm connectors inspect destination-starburst-galaxy - Starburst Galaxy connector manual

SYNOPSIS
  pm connectors inspect destination-starburst-galaxy
  pm connectors inspect destination-starburst-galaxy --json
  pm credentials add <name> --connector destination-starburst-galaxy [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Starburst Galaxy catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/starburst-galaxy.svg
  source: official
  review_status: official_verified
  review_url: https://docs.starburst.io/starburst-galaxy/

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  Starburst Galaxy documentation: https://docs.starburst.io/starburst-galaxy/

CONFIGURATION
  accept_terms (boolean) required: You must agree to the Starburst Galaxy <a href="https://www.starburst.io/terms/">terms & conditions</a> to use this connector.
  catalog (string) required: Name of the Starburst Galaxy Amazon S3 catalog.
  catalog_schema (string): The default Starburst Galaxy Amazon S3 catalog schema where tables are written to if the source does not specify a namespace. Defaults to "public".
  password (string) required secret: Starburst Galaxy password for the specified user.
  port (string): Starburst Galaxy cluster port.
  purge_staging_table (boolean): Defaults to 'true'. Switch to 'false' for debugging purposes.
  server_hostname (string) required: Starburst Galaxy cluster hostname.
  staging_object_store (object) required: Temporary storage on which temporary Iceberg table is created.
  username (string) required: Starburst Galaxy user.
  secret fields: password, staging_object_store.s3_access_key_id, staging_object_store.s3_secret_access_key

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-starburst-galaxy

  # Inspect as JSON
  pm connectors inspect destination-starburst-galaxy --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Starburst Galaxy documentation: https://docs.starburst.io/starburst-galaxy/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
