# pm connectors inspect source-couchbase

```text
NAME
  pm connectors inspect source-couchbase - Couchbase connector manual

SYNOPSIS
  pm connectors inspect source-couchbase
  pm connectors inspect source-couchbase --json
  pm credentials add <name> --connector source-couchbase [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Couchbase catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/couchbase.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  Couchbase SQL++ reference: https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html
  Couchbase authentication: https://docs.couchbase.com/server/current/learn/security/authentication.html

CONFIGURATION
  bucket (string) required: The name of the bucket to sync data from
  connection_string (string) required: The connection string for the Couchbase server (e.g., couchbase://localhost or couchbases://example.com)
  password (string) required secret: The password to use for authentication
  start_date (string): The date from which you'd like to replicate data for incremental streams, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated. If not set, ...
  username (string) required: The username to use for authentication
  secret fields: password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-couchbase

  # Inspect as JSON
  pm connectors inspect source-couchbase --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Couchbase SQL++ reference: https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html
  Couchbase authentication: https://docs.couchbase.com/server/current/learn/security/authentication.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
