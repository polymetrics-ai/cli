# pm connectors inspect k6-cloud

```text
NAME
  pm connectors inspect k6-cloud - k6 Cloud connector manual

SYNOPSIS
  pm connectors inspect k6-cloud
  pm connectors inspect k6-cloud --json
  pm credentials add <name> --connector k6-cloud [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads k6 Cloud organizations, projects, and load tests through the k6 Cloud REST API.

ICON
  id: k6cloud
  asset: icons/k6cloud.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://k6.io/docs/cloud/cloud-reference/cloud-rest-api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  api_token (secret) (required)

ETL STREAMS
  organizations:
    primary key: id
    fields: billing_address(string), billing_country(string), billing_email(string), created(string), description(string), id(integer), is_default(boolean), is_saml_org(boolean), name(string), owner_id(integer), updated(string), vat_number(string)
  k6_tests:
    primary key: id
    fields: created(string), id(integer), last_test_run_id(string), name(string), project_id(integer), script(string), test_run_ids(array), updated(string), user_id(integer)
  projects:
    primary key: id
    fields: created(string), description(string), id(integer), is_default(boolean), name(string), organization_id(integer), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external k6 Cloud API read of organizations, projects, and load tests
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect k6-cloud

  # Inspect as structured JSON
  pm connectors inspect k6-cloud --json

AGENT WORKFLOW
  - Run pm connectors inspect k6-cloud before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
