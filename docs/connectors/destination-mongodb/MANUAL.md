# pm connectors inspect destination-mongodb

```text
NAME
  pm connectors inspect destination-mongodb - MongoDB connector manual

SYNOPSIS
  pm connectors inspect destination-mongodb
  pm connectors inspect destination-mongodb --json
  pm credentials add <name> --connector destination-mongodb [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MongoDB catalog connector for https://docs.airbyte.com/integrations/destinations/mongodb. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-mongodb:0.2.0 (metadata only; not executed)

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
  MongoDB documentation: https://www.mongodb.com/docs/
  Authentication: https://www.mongodb.com/docs/manual/core/authentication/
  Role-based access control: https://www.mongodb.com/docs/manual/core/authorization/
  Release Notes: https://www.mongodb.com/docs/manual/release-notes/
  MongoDB Status: https://status.mongodb.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/mongodb

CONFIGURATION
  auth_type (object) required: Authorization type.
  database (string) required: Name of the database.
  instance_type (object): MongoDb instance to connect to. For MongoDB Atlas and Replica Set TLS connection is used by default.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  secret fields: auth_type.password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/mongodb

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-mongodb

  # Inspect as JSON
  pm connectors inspect destination-mongodb --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MongoDB documentation: https://docs.airbyte.com/integrations/destinations/mongodb

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
