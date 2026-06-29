# pm connectors inspect destination-elasticsearch

```text
NAME
  pm connectors inspect destination-elasticsearch - ElasticSearch connector manual

SYNOPSIS
  pm connectors inspect destination-elasticsearch
  pm connectors inspect destination-elasticsearch --json
  pm credentials add <name> --connector destination-elasticsearch [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ElasticSearch catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/elasticsearch.svg
  source: official
  review_status: official_verified
  review_url: https://www.elastic.co/docs/reference/elasticsearch

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
  ElasticSearch documentation: https://www.elastic.co/docs/reference/elasticsearch

CONFIGURATION
  authenticationMethod (object): The type of authentication to be used
  ca_certificate (string) secret: CA certificate
  endpoint (string) required: The full url of the Elasticsearch server
  pathPrefix (string): The Path Prefix of the Elasticsearch server
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  upsert (boolean): If a primary key identifier is defined in the source, an upsert will be performed using the primary key value as the elasticsearch doc id. Does not support composite primary keys.
  secret fields: authenticationMethod.apiKeySecret, authenticationMethod.password, ca_certificate, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-elasticsearch

  # Inspect as JSON
  pm connectors inspect destination-elasticsearch --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  ElasticSearch documentation: https://www.elastic.co/docs/reference/elasticsearch

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
