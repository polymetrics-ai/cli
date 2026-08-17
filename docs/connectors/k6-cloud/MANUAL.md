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
  api_token (secret)

ETL STREAMS
  organizations:
    primary key: id
    fields: billing_address(), billing_country(), billing_email(), created(), description(), id(), is_default(), is_saml_org(), name(), owner_id(), updated(), vat_number()
  k6_tests:
    primary key: id
    fields: created(), id(), last_test_run_id(), name(), project_id(), script(), test_run_ids(), updated(), user_id()
  projects:
    primary key: id
    fields: created(), description(), id(), is_default(), name(), organization_id(), updated()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
