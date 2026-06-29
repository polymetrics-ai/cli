# pm connectors inspect destination-salesforce

```text
NAME
  pm connectors inspect destination-salesforce - Salesforce connector manual

SYNOPSIS
  pm connectors inspect destination-salesforce
  pm connectors inspect destination-salesforce --json
  pm credentials add <name> --connector destination-salesforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Salesforce catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/salesforce.svg
  source: official
  review_status: official_verified
  review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: certified

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
  Salesforce documentation: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

CONFIGURATION
  auth_type (string) required
  client_id (string) required: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client ID</a>.
  client_secret (string) required secret: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client secret</a>.
  is_sandbox (boolean) required: Toggle if you're using a <a href="https://help.salesforce.com/s/articleView?id=sf.deploy_sandboxes_parent.htm&type=5">Salesforce Sandbox</a>.
  object_storage_config (object)
  refresh_token (string) required secret: manual intervention needed
  secret fields: client_secret, object_storage_config.access_key_id, object_storage_config.secret_access_key, refresh_token

SYNC MODES
  supported sync modes: append
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-salesforce

  # Inspect as JSON
  pm connectors inspect destination-salesforce --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Salesforce documentation: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
