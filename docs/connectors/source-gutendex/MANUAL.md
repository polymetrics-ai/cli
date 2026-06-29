# pm connectors inspect source-gutendex

```text
NAME
  pm connectors inspect source-gutendex - Gutendex connector manual

SYNOPSIS
  pm connectors inspect source-gutendex
  pm connectors inspect source-gutendex --json
  pm credentials add <name> --connector source-gutendex [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Gutendex catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/source-gutendex.svg
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
  Gutendex documentation: https://gutendex.com/

CONFIGURATION
  author_year_end (string): (Optional) Defines the maximum birth year of the authors. Books by authors born after the end year will not be returned. Supports both positive (CE) or negative (BCE) integer va...
  author_year_start (string): (Optional) Defines the minimum birth year of the authors. Books by authors born prior to the start year will not be returned. Supports both positive (CE) or negative (BCE) integ...
  copyright (string): (Optional) Use this to find books with a certain copyright status - true for books with existing copyrights, false for books in the public domain in the USA, or null for books w...
  languages (string): (Optional) Use this to find books in any of a list of languages. They must be comma-separated, two-character language codes.
  search (string): (Optional) Use this to search author names and book titles with given words. They must be separated by a space (i.e. %20 in URL-encoded format) and are case-insensitive.
  sort (string): (Optional) Use this to sort books - ascending for Project Gutenberg ID numbers from lowest to highest, descending for IDs highest to lowest, or popular (the default) for most po...
  topic (string): (Optional) Use this to search for a case-insensitive key-phrase in books' bookshelves or subjects.

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gutendex

  # Inspect as JSON
  pm connectors inspect source-gutendex --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Gutendex documentation: https://gutendex.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
