# pm connectors inspect destination-hubspot

```text
NAME
  pm connectors inspect destination-hubspot - HubSpot connector manual

SYNOPSIS
  pm connectors inspect destination-hubspot
  pm connectors inspect destination-hubspot --json
  pm credentials add <name> --connector destination-hubspot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  HubSpot catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/hubspot.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.hubspot.com/docs/api/overview

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
  HubSpot API documentation: https://developers.hubspot.com/docs/api/overview
  OAuth: https://developers.hubspot.com/docs/api/oauth-quickstart-guide
  HubSpot Developer Changelog: https://developers.hubspot.com/changelog
  Rate limits: https://developers.hubspot.com/docs/api/usage-details
  HubSpot Status: https://status.hubspot.com/

CONFIGURATION
  credentials (object) required: Choose how to authenticate to HubSpot.
  object_storage_config (object)
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, object_storage_config.access_key_id, object_storage_config.secret_access_key

SYNC MODES
  supported sync modes: append
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-hubspot

  # Inspect as JSON
  pm connectors inspect destination-hubspot --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  HubSpot API documentation: https://developers.hubspot.com/docs/api/overview
  OAuth: https://developers.hubspot.com/docs/api/oauth-quickstart-guide
  HubSpot Developer Changelog: https://developers.hubspot.com/changelog
  Rate limits: https://developers.hubspot.com/docs/api/usage-details
  HubSpot Status: https://status.hubspot.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
