# pm connectors inspect source-serpstat

```text
NAME
  pm connectors inspect source-serpstat - Serpstat connector manual

SYNOPSIS
  pm connectors inspect source-serpstat
  pm connectors inspect source-serpstat --json
  pm credentials add <name> --connector source-serpstat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Serpstat catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/serpstat.svg
  source: upstream_registry
  review_status: upstream_seeded

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
  Serpstat documentation: https://api-docs.serpstat.com/

CONFIGURATION
  api_key (string) required secret: Serpstat API key can be found here: https://serpstat.com/users/profile/
  domain (string): The domain name to get data for (ex. serpstat.com)
  domains (array): The list of domains that will be used in streams that support batch operations
  filter_by (string): The field name by which the results should be filtered. Filtering the results will result in fewer API credits spent. Each stream has different filtering options. See https://se...
  filter_value (string): The value of the field to filter by. Each stream has different filtering options. See https://serpstat.com/api/ for more details.
  page_size (integer): The number of data rows per page to be returned. Each data row can contain multiple data points. The max value is 1000. Reducing the size of the page will result in fewer API cr...
  pages_to_fetch (integer): The number of pages that should be fetched. All results will be obtained if left blank. Reducing the number of pages will result in fewer API credits spent.
  region_id (string): The ID of a region to get data from in the form of a two-letter country code prepended with the g_ prefix. See the list of supported region IDs here: https://serpstat.com/api/66...
  sort_by (string): The field name by which the results should be sorted. Each stream has different sorting options. See https://serpstat.com/api/ for more details.
  sort_value (string): The value of the field to sort by. Each stream has different sorting options. See https://serpstat.com/api/ for more details.
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
  pm connectors inspect source-serpstat

  # Inspect as JSON
  pm connectors inspect source-serpstat --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Serpstat documentation: https://api-docs.serpstat.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
