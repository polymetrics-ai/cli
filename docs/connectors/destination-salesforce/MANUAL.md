# pm connectors inspect destination-salesforce

```text
NAME
  pm connectors inspect destination-salesforce - Salesforce connector manual

SYNOPSIS
  pm connectors inspect destination-salesforce
  pm connectors inspect destination-salesforce --json
  pm credentials add <name> --connector destination-salesforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Salesforce catalog connector for https://docs.airbyte.com/integrations/enterprise-connectors/destination-salesforce. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-salesforce:0.0.8 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/enterprise-connectors/destination-salesforce

CONFIGURATION
  auth_type (string) required
  client_id (string) required: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client ID</a>.
  client_secret (string) required secret: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client secret</a>.
  is_sandbox (boolean) required: Toggle if you're using a <a href="https://help.salesforce.com/s/articleView?id=sf.deploy_sandboxes_parent.htm&type=5">Salesforce Sandbox</a>.
  object_storage_config (object)
  refresh_token (string) required secret: Enter your application's <a href="https://developer.salesforce.com/docs/atlas.en-us.mobile_sdk.meta/mobile_sdk/oauth_refresh_token_flow.htm">Salesforce Refresh Token</a> used fo...
  secret fields: client_secret, object_storage_config.access_key_id, object_storage_config.secret_access_key, refresh_token

SYNC MODES
  supported sync modes: append
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/enterprise-connectors/destination-salesforce

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
  Salesforce documentation: https://docs.airbyte.com/integrations/enterprise-connectors/destination-salesforce

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
