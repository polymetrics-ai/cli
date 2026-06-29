# pm connectors inspect source-faker

```text
NAME
  pm connectors inspect source-faker - Sample Data connector manual

SYNOPSIS
  pm connectors inspect source-faker
  pm connectors inspect source-faker --json
  pm credentials add <name> --connector source-faker [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Sample Data catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
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
  family: custom_go_port
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Python Faker Library Documentation: https://faker.readthedocs.io/en/master/
  Faker Changelog: https://github.com/joke2k/faker/blob/master/CHANGELOG.md

CONFIGURATION
  always_updated (boolean): Should the updated_at values for every record be new each sync? Setting this to false will case the source to stop emitting records after COUNT records have been emitted.
  count (integer): How many users should be generated in total. The purchases table will be scaled to match, with 10 purchases created per 10 users. This setting does not apply to the products str...
  parallelism (integer): How many parallel workers should we use to generate fake data? Choose a value equal to the number of CPUs you will allocate to this source.
  records_per_slice (integer): How many fake records will be in each page (stream slice), before a state message is emitted?
  seed (integer): Manually control the faker random seed to return the same values on subsequent runs (leave -1 for random)

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-faker

  # Inspect as JSON
  pm connectors inspect source-faker --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Python Faker Library Documentation: https://faker.readthedocs.io/en/master/
  Faker Changelog: https://github.com/joke2k/faker/blob/master/CHANGELOG.md

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
