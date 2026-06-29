# pm connectors inspect destination-deepset

```text
NAME
  pm connectors inspect destination-deepset - Deepset connector manual

SYNOPSIS
  pm connectors inspect destination-deepset
  pm connectors inspect destination-deepset --json
  pm credentials add <name> --connector destination-deepset [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Deepset catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/deepset.svg
  source: upstream_registry
  review_status: upstream_seeded

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
  Deepset documentation: https://docs.cloud.deepset.ai/docs/getting-started

CONFIGURATION
  api_key (string) required secret: Your deepset cloud API key
  base_url (string): URL of deepset Cloud API (e.g. https://api.cloud.deepset.ai, https://api.us.deepset.ai, etc). Defaults to https://api.cloud.deepset.ai.
  retries (number): Number of times to retry an action before giving up.
  workspace (string) required: Name of workspace to which to sync the data.
  secret fields: api_key

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-deepset

  # Inspect as JSON
  pm connectors inspect destination-deepset --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Deepset documentation: https://docs.cloud.deepset.ai/docs/getting-started

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
