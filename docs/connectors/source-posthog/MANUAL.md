# pm connectors inspect source-posthog

```text
NAME
  pm connectors inspect source-posthog - PostHog connector manual

SYNOPSIS
  pm connectors inspect source-posthog
  pm connectors inspect source-posthog --json
  pm credentials add <name> --connector source-posthog [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  PostHog catalog connector for https://docs.airbyte.com/integrations/sources/posthog. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-posthog:1.1.25 (metadata only; not executed)

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
  family: custom_go_port
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/posthog

CONFIGURATION
  api_key (string) required secret: API Key. See the <a href="https://docs.airbyte.com/integrations/sources/posthog">docs</a> for information on how to generate this key.
  base_url (string): Base PostHog url. Defaults to PostHog Cloud (https://app.posthog.com).
  events_time_step (integer): Set lower value in case of failing long running sync of events stream.
  start_date (string) required: The date from which you'd like to replicate the data. Any data before this date will not be replicated.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/posthog

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-posthog

  # Inspect as JSON
  pm connectors inspect source-posthog --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  PostHog documentation: https://docs.airbyte.com/integrations/sources/posthog

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
