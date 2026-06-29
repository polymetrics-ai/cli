# pm connectors inspect destination-firebolt

```text
NAME
  pm connectors inspect destination-firebolt - Firebolt connector manual

SYNOPSIS
  pm connectors inspect destination-firebolt
  pm connectors inspect destination-firebolt --json
  pm credentials add <name> --connector destination-firebolt [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Firebolt catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/firebolt.svg
  source: official
  review_status: official_verified
  review_url: https://docs.firebolt.io/overview

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
  Firebolt documentation: https://docs.firebolt.io/overview

CONFIGURATION
  account (string) required: Firebolt account to login.
  client_id (string) required: Firebolt service account ID.
  client_secret (string) required secret: Firebolt secret, corresponding to the service account ID.
  database (string) required: The database to connect to.
  engine (string) required: Engine name to connect to.
  host (string): The host name of your Firebolt database.
  loading_method (object): Loading method used to select the way data will be uploaded to Firebolt
  secret fields: client_secret, loading_method.aws_key_id, loading_method.aws_key_secret

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-firebolt

  # Inspect as JSON
  pm connectors inspect destination-firebolt --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Firebolt documentation: https://docs.firebolt.io/overview

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
