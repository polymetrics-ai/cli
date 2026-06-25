# pm connectors inspect destination-redis

```text
NAME
  pm connectors inspect destination-redis - Redis connector manual

SYNOPSIS
  pm connectors inspect destination-redis
  pm connectors inspect destination-redis --json
  pm credentials add <name> --connector destination-redis [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Redis catalog connector for https://docs.airbyte.com/integrations/destinations/redis. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-redis:0.1.4 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/redis

CONFIGURATION
  cache_type (string) required: Redis cache type to store data in.
  host (string) required: Redis host to connect to.
  password (string) secret: Password associated with Redis.
  port (integer) required: Port of Redis.
  ssl (boolean): Indicates whether SSL encryption protocol will be used to connect to Redis. It is recommended to use SSL connection if possible.
  ssl_mode (object): SSL connection modes. <li><b>verify-full</b> - This is the most secure mode. Always require encryption and verifies the identity of the source database server
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: Username associated with Redis.
  secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/redis

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-redis

  # Inspect as JSON
  pm connectors inspect destination-redis --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Redis documentation: https://docs.airbyte.com/integrations/destinations/redis

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
