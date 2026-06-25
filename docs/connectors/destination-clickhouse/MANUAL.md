# pm connectors inspect destination-clickhouse

```text
NAME
  pm connectors inspect destination-clickhouse - ClickHouse connector manual

SYNOPSIS
  pm connectors inspect destination-clickhouse
  pm connectors inspect destination-clickhouse --json
  pm credentials add <name> --connector destination-clickhouse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ClickHouse catalog connector for https://docs.airbyte.com/integrations/destinations/clickhouse. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-clickhouse:2.1.24 (metadata only; not executed)

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
  ClickHouse documentation: https://clickhouse.com/docs
  SQL reference: https://clickhouse.com/docs/en/sql-reference
  User authentication: https://clickhouse.com/docs/en/operations/access-rights
  Changelog: https://clickhouse.com/docs/whats-new/changelog
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/clickhouse

CONFIGURATION
  database (string) required: Name of the database.
  enable_json (boolean): Use the JSON type for Object fields. If disabled, the JSON will be converted to a string.
  host (string) required: Hostname of the database.
  password (string) required secret: Password associated with the username.
  port (string) required: HTTP port of the database. Default(s) HTTP: 8123 — HTTPS: 8443
  protocol (string) required: Protocol for the database connection string.
  record_window_size (integer): Warning: Tuning this parameter can impact the performances. The maximum number of records that should be written to a batch. The batch size limit is still limited to 70 Mb
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: Username to use to access the database.
  secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/clickhouse

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-clickhouse

  # Inspect as JSON
  pm connectors inspect destination-clickhouse --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  ClickHouse documentation: https://docs.airbyte.com/integrations/destinations/clickhouse

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
