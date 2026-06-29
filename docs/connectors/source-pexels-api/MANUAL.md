# pm connectors inspect source-pexels-api

```text
NAME
  pm connectors inspect source-pexels-api - Pexels API connector manual

SYNOPSIS
  pm connectors inspect source-pexels-api
  pm connectors inspect source-pexels-api --json
  pm credentials add <name> --connector source-pexels-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pexels API catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pexels.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.pexels.com/api/documentation/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
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
  family: declarative_http_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Pexels API documentation: https://www.pexels.com/api/documentation/

CONFIGURATION
  api_key (string) required secret: API key is required to access pexels api, For getting your's goto https://www.pexels.com/api/documentation and create account for free.
  color (string): Optional, Desired photo color. Supported colors red, orange, yellow, green, turquoise, blue, violet, pink, brown, black, gray, white or any hexidecimal color code.
  locale (string): Optional, The locale of the search you are performing. The current supported locales are 'en-US' 'pt-BR' 'es-ES' 'ca-ES' 'de-DE' 'it-IT' 'fr-FR' 'sv-SE' 'id-ID' 'pl-PL' 'ja-JP' ...
  orientation (string): Optional, Desired photo orientation. The current supported orientations are landscape, portrait or square
  query (string) required: Optional, the search query, Example Ocean, Tigers, Pears, etc.
  size (string): Optional, Minimum photo size. The current supported sizes are large(24MP), medium(12MP) or small(4MP).
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pexels-api

  # Inspect as JSON
  pm connectors inspect source-pexels-api --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Pexels API documentation: https://www.pexels.com/api/documentation/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
